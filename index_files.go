package main

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func prepareBooks() ([]Book, error) {

	file, err := os.Open("mehaz/kitap.csv")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	r := csv.NewReader(file)
	r.Comma = ';'
	r.Comment = '#'

	books := make([]Book, 0)

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

	return books, nil
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

func reindexAllFiles() {
	fileInfos, err := ioutil.ReadDir("books")
	if err != nil {
		log.Printf("opening books directory failed.")
		return
	}

	bookId := 0
	for _, file := range fileInfos {
		if filepath.Ext(file.Name()) == ".json" {
			book, err := loadBookMeta(file.Name())
			if err != nil {
				log.Printf("loading file meta from json file:%s faied\n", err)
				continue
			}
			book.Id = uint32(bookId)
			log.Println(book)
			indexBook(book)
			bookId++

			//store payload data in cdb file
			//ProcessPayloadFile(book.Hash)
		}
	}
}

func indexBook(book Book) {
	pages, err := getPages(book)
	if err != nil {
		fmt.Println(err)
	}

	book.NumPages = uint32(len(pages))

	//log.Println(pages)

	booksMap[uint32(book.Id)] = book

	for _, page := range pages {
		docId := idx.Add(page.Content, book.Category)
		page.Id = docId
		page.BookId = book.Id
		pagesMap[docId] = page
	}

	//processPayload(book.Hash)
	//syncPayload()
}

func gobEncode(tokens map[string][][4]int) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(tokens); err != nil {
		log.Fatal(err)
	}

	return buf.Bytes()
}
