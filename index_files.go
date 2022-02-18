package main

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func prepareBooks(path string) ([]Book, error) {

	cvsFile := filepath.Join(path, "books.csv")
	books := make([]Book, 0)

	// if folder has a "books.csv" file use it index files
	// else scan path and subfolders for .pdf files
	if ok, _ := exists(cvsFile); ok {

		file, err := os.Open(cvsFile)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		r := csv.NewReader(file)
		r.Comma = ';'
		r.Comment = '#'

		for {
			record, err := r.Read()
			log.Println(record)
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			book := Book{}
			// remove file extention
			lastIndex := strings.LastIndex(record[0], ".")
			book.Title = record[0][0:lastIndex]

			categories := strings.Split(record[1], ",")

			book.Category = append(book.Category, categories...)

			hash, err := preparePdfFile("mehaz/" + record[0])
			if err != nil {
				log.Println(err)
				continue
				//return nil, err
			}

			book.Hash = hash
			books = append(books, book)

			processPdfFile(book)

			// save book struct as json file
			saveBookMeta(book)
		}

	} else {
		books = indexDirectory(path)
	}

	return books, nil
}

func indexDirectory(path string) []Book {

	books := make([]Book, 0)

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err.Error())
		} else {
			ext := filepath.Ext(info.Name())
			if strings.ToLower(ext) == ".pdf" {
				book := Book{}
				name := strings.TrimSuffix(info.Name(), ext)

				book.Title = name
				hash, err := preparePdfFile(path)
				if err != nil {
					log.Println(err)
				}
				book.Hash = hash
				books = append(books, book)

				processPdfFile(book)
				saveBookMeta(book)

				log.Printf("file name: %s\n", info.Name())
			}
		}

		return nil
	})

	return books
}

func saveBookMeta(book Book) error {

	bookJson, err := json.Marshal(book)
	if err != nil {
		return err
	}

	file, err := os.Create("books/" + book.Hash + ".json")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(bookJson)
	if err != nil {
		return err
	}

	return nil
}

func loadBookMeta(filename string) (Book, error) {

	book := Book{}

	file, err := os.Open("books/" + filename)
	if err != nil {
		return book, err
	}
	defer file.Close()

	bookJson, err := ioutil.ReadAll(file)
	if err != nil {
		return book, err
	}

	err = json.Unmarshal(bookJson, &book)
	if err != nil {
		return book, err
	}

	return book, err
}

func indexBook(book Book) error {
	pages, err := getPages(book)
	if err != nil {
		log.Println(err)
		return err
	}

	book.NumPages = uint32(len(pages))

	//log.Println(pages)

	bookIndex.booksMap[uint32(book.Id)] = book

	for _, page := range pages {
		docId := bookIndex.idx.Add(page.Content, book.Category)
		page.Id = docId
		page.BookId = book.Id
		bookIndex.pagesMap[docId] = page
	}

	//processPayload(book.Hash)
	//syncPayload()

	return nil
}
