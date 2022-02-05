package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/aydink/inverted"
)

var booksMap map[uint32]Book
var pagesMap map[uint32]Page

var idx *inverted.InvertedIndex
var simpleHighlighter inverted.SimpleHighlighter
var spanHighlighter inverted.SpanHighlighter

var payloadStore *CdbStore

var turkishAnalyzer *inverted.SimpleAnalyzer

func buildIndex() {

	turkishAnalyzer := inverted.NewSimpleAnalyzer(inverted.NewSimpleTokenizer())
	turkishAnalyzer.AddTokenFilter(inverted.NewTurkishLowercaseFilter())
	turkishAnalyzer.AddTokenFilter(inverted.NewTurkishAccentFilter())
	turkishAnalyzer.AddTokenFilter(inverted.NewTurkishStemFilter())
	//analyzer.AddTokenFilter(NewEnglishStemFilter())

	simpleHighlighter = inverted.NewSimpleHighlighter(turkishAnalyzer)
	spanHighlighter = inverted.NewSpanHighlighter(turkishAnalyzer)

	if *flagInMermory {
		idx = inverted.NewInvertedIndex(turkishAnalyzer)

		indexFiles()
		idx.UpdateAvgFieldLen()
		idx.BuildCategoryBitmap()

		idx.MarshalIndex()

		serializeBooks(booksMap)
		serializePages(pagesMap)
	} else {
		idx = inverted.NewInvertedIndexFromFile(turkishAnalyzer, false)
		booksMap = deserializeBooks()
		//fmt.Println(booksMap)
		pagesMap = deserializePages()
		//fmt.Println(pagesMap)
	}

	buildPayloadDatabase()
	/*
		idx = inverted.NewInvertedIndexFromFile(analyzer, false)
		booksMap = deserializeBooks()
		//fmt.Println(booksMap)
		pagesMap = deserializePages()
		//fmt.Println(pagesMap)
	*/
}

func buildPayloadDatabase() {

	var err error

	if *flagBuildPayload {

		payloadStore, err = NewCdbStore()
		if err != nil {
			log.Println("Failed to create cdb file")
			return
		}

		payloadStore.BuildDatabase()
		payloadStore.Freeze()
	} else {

		payloadStore, err = OpenCdbStore()
		if err != nil {
			log.Println("Failed to open cdb file")
			return
		}
	}
}

func indexFiles() {

	if *flagRebuild {
		books, err := prepareBooks()
		if err != nil {
			fmt.Println("Failed to load book list csv file", err)
			return
		}

		for i, book := range books {
			book.Id = uint32(i)
			booksMap[book.Id] = book
			indexBook(book)
		}
	} else {
		reindexAllFiles()
	}
}

func cleanUpBeforeExit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			fmt.Println(sig.String(), "Ctrl-C captured")
			payloadStore.reader.Close()
			os.Exit(0)
		}
	}()
}

func GetBook(hash string) Book {
	for _, v := range booksMap {
		if v.Hash == hash {
			return v
		}
	}

	return Book{}
}

var flagRebuild *bool
var flagBuildPayload *bool
var flagInMermory *bool

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flagRebuild = flag.Bool("rebuild", false, "rebuild index form scratch using csv file")
	flagBuildPayload = flag.Bool("payload", false, "rebuild payload cdb file form scratch")
	flagInMermory = flag.Bool("inmemory", true, "create an inmermoy index or open from disk")

	flag.Parse()

	fmt.Println(*flagRebuild)
	fmt.Println(*flagBuildPayload)

	booksMap = make(map[uint32]Book)
	pagesMap = make(map[uint32]Page)

	//go printMemUsage()

	// capture Ctrl-C exit event
	cleanUpBeforeExit()

	// build fulltext index
	buildIndex()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/test", test)
	http.HandleFunc("/search/", searchHandler)
	http.HandleFunc("/page", pageHandler)
	http.HandleFunc("/image", imageHandler)
	http.HandleFunc("/download/", downloadHandler)
	http.HandleFunc("/stats", tokenStatHandler)
	http.HandleFunc("/books", booksHandler)
	http.HandleFunc("/api/addbook", uploadHandler)
	http.HandleFunc("/api/payloads", payloadHandler)
	http.HandleFunc("/reset", resetIndexHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	fmt.Println("--------------------------------------------------")
	fmt.Println("Arama motorunu kullanmak için tarayıcı ile http://127.0.0.1:8080 adresine gidin")

	openBrowser("http://127.0.0.1:8080/")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}

}
