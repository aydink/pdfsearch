package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type HitResult struct {
	Book   Book
	Page   Page
	HlText string
}

// getFullFilterName return full name of the filter
// eg. "year": "Yıl", "genre": "Türü"
func getFullFilterName(key string) string {

	filterFullNames := map[string]string{
		"year":     "Basım yılı",
		"genre":    "Türü",
		"category": "Kategori",
	}

	if value, found := filterFullNames[key]; found {
		return value
	}
	return key
}

func getFilters(v url.Values) [][3]string {
	filterNames := []string{"genre", "type", "category"}

	filters := make([][3]string, 0)

	for _, name := range filterNames {
		if v.Get(name) != "" {
			filters = append(filters, [3]string{name, getFullFilterName(name), v.Get(name)})
			//filters = v["category"]
		}
	}
	return filters
}

func getFilters2(v url.Values) []string {

	filters := make([]string, 0)
	s := v.Get("category")

	if s != "" {
		filters = strings.Split(s, ":")
	}

	return filters
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")

	if err != nil {
		fmt.Fprintf(w, "Hata: %s!", err)
	}
	t.Execute(w, nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {

	v := r.URL.Query()
	filters := getFilters2(v)

	//t, err := template.ParseFiles("templates/search.html")
	//t := template.Must(template.New("").Funcs(funcMap).ParseFiles("templates/search.html", "templates/partial_facet.html", "templates/partial_pagination.html", "templates/partial_definition.html"))
	t := template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))

	q := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")
	//searchType := r.URL.Query().Get("w")

	start := r.URL.Query().Get("start")
	startInt, err := strconv.Atoi(start)
	//	fmt.Println("start:", startInt)

	if err != nil {
		//fmt.Println("error parsing 'start' parameter")
		startInt = 0
	}

	if startInt < 0 {
		startInt = 0
	}

	templateName := "search"

	hits := bookIndex.idx.Search_Mixed_v2(q)

	if len(category) > 0 {
		hits = bookIndex.idx.FacetFilter(hits, category)
	}

	data := make(map[string]interface{})
	data["q"] = q
	data["categoryFacet"] = bookIndex.idx.GetFacetCounts(hits)

	totalHits := len(hits)
	data["TotalHits"] = totalHits

	data["numPages"], data["pages"] = Paginate(startInt, 10, len(hits))
	data["filters"] = filters

	if startInt < totalHits {
		if (startInt + 10) < totalHits {
			hits = hits[startInt : startInt+10]
		} else {
			hits = hits[startInt:]
		}
	}

	data["currentPage"] = startInt

	/*
		if len(hits) > 10 {
			hits = hits[0:10]
		}
	*/
	hitResults := make([]HitResult, 0)

	for _, hit := range hits {

		//log.Println(hit)

		result := HitResult{}
		result.Page = bookIndex.pagesMap[hit.DocId]
		result.Book = bookIndex.booksMap[result.Page.BookId]
		//result.HlText = simpleHighlighter.Highlight("<b>", "</b>", pagesMap[hit.DocId].Content, q)
		result.HlText = bookIndex.spanHighlighter.Highlight(bookIndex.pagesMap[hit.DocId].Content, 200, q)

		hitResults = append(hitResults, result)

	}

	data["hits"] = hitResults

	err = t.ExecuteTemplate(w, templateName, data)
	if err != nil {
		fmt.Println(err)
	}
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))

	q := r.URL.Query().Get("q")
	page := r.URL.Query().Get("page")
	pageInt, err := strconv.Atoi(page)

	if err != nil {
		pageInt = 0
	}

	curPage := bookIndex.pagesMap[uint32(pageInt)]
	curBook := bookIndex.booksMap[curPage.BookId]
	hash := curBook.Hash

	image := hash + "-" + strconv.Itoa(int(curPage.PageNumber))

	createImage(image)

	data := make(map[string]interface{})
	data["q"] = q
	data["title"] = curBook.Title
	data["image"] = image
	data["hash"] = hash
	data["numPages"] = curBook.NumPages
	data["curPage"] = curPage.PageNumber
	data["pageId"] = pageInt

	//fmt.Printf("%+v", data)
	t.ExecuteTemplate(w, "document", data)
}

func payloadHandler(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	q := r.URL.Query().Get("q")

	w.Header().Set("Content-Type", "application/json")
	//fmt.Fprint(w, GetTokenPositions(page, q))
	fmt.Fprint(w, bookIndex.payloadStore.GetTokenPositions(page, q))
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("page")
	parts := strings.Split(query, "-")
	hash := parts[0]
	page := parts[1]

	createImage(query)

	http.ServeFile(w, r, "static/images/"+hash+"-"+page+".png")
}

func createImage(query string) {

	parts := strings.Split(query, "-")
	hash := parts[0]
	page := parts[1]

	//fmt.Println("hash:", hash, "page:", page, "file:", fileMap[hash])

	if _, err := os.Stat("static/images/" + hash + "-" + page + ".png"); os.IsNotExist(err) {
		_, err := exec.Command("pdftocairo", "-png", "-singlefile", "-f", page, "-l", page, "books/"+hash+".pdf", "static/images/"+hash+"-"+page).Output()
		if err != nil {
			log.Println(err)
		}
	} else {
		//fmt.Println("-----------------------", "using cashed image")
	}
}

// PDF file handler
// send pdf file and sets a proper title
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("book")
	if len(hash) > 32 {
		hash = hash[:32]
	}

	if len(hash) < 32 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Geçersiz bir istekte bulundunuz.")
		log.Printf("download pdf, invalid hash value:%s", hash)
		return
	}

	// check if user wants to download file
	force := r.URL.Query().Get("force")

	file, err := os.Open("books/" + hash + ".pdf")
	defer file.Close()
	if err != nil {
		log.Printf("failed to serve pdf file:%s", "books/"+hash+".pdf")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Geçersiz bir istekte bulundunuz.")
		return
	}

	book := bookIndex.GetBook(hash)
	name := book.Title

	// if there is an explicit url prameter "force=true" then force browser to download not try to display the pdf file
	if force == "true" {
		w.Header().Set("Content-Disposition", "attachment; filename="+name+".pdf")
	}

	w.Header().Set("Content-Type", "application/pdf")
	io.Copy(w, file)
}

// send token statistics
func tokenStatHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Disposition", "attachment; filename=token_stats.txt")
	w.Header().Set("Content-Type", "text/plain")

	for _, v := range bookIndex.idx.TokenStats() {
		fmt.Fprintf(w, "%s\t%d\n", v.Name, v.Count)
	}
}

func booksHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))

	sorted := make([]Book, len(bookIndex.booksMap))
	for _, book := range bookIndex.booksMap {
		sorted = append(sorted, book)
	}

	sort.Sort(byBookTitle(sorted))

	data := make(map[string]interface{})
	data["q"] = ""
	data["books"] = sorted
	//log.Println(data)

	t.ExecuteTemplate(w, "books", data)
}

// send token statistics
func test(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, bookIndex.booksMap)
}
