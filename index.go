package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/aydink/inverted"
)

type BookIndex struct {
	idx               *inverted.InvertedIndex
	simpleHighlighter *inverted.SimpleHighlighter
	spanHighlighter   *inverted.SpanHighlighter
	//turkishAnalyzer   *inverted.SimpleAnalyzer

	payloadStore *CdbStore

	booksMap map[uint32]Book
	pagesMap map[uint32]Page
}

func NewBookIndex() *BookIndex {

	bookIndex := &BookIndex{}

	turkishAnalyzer := inverted.NewSimpleAnalyzer(inverted.NewSimpleTokenizer())
	turkishAnalyzer.AddTokenFilter(inverted.NewTurkishLowercaseFilter())
	turkishAnalyzer.AddTokenFilter(inverted.NewTurkishAccentFilter())
	turkishAnalyzer.AddTokenFilter(inverted.NewTurkishStemFilter())

	bookIndex.idx = inverted.NewInvertedIndex(turkishAnalyzer)

	bookIndex.simpleHighlighter = inverted.NewSimpleHighlighter(turkishAnalyzer)
	bookIndex.spanHighlighter = inverted.NewSpanHighlighter(turkishAnalyzer)

	bookIndex.booksMap = make(map[uint32]Book)
	bookIndex.pagesMap = make(map[uint32]Page)

	return bookIndex
}

func OpenBookIndex() *BookIndex {

	bookIndex := &BookIndex{}

	turkishAnalyzer := inverted.NewSimpleAnalyzer(inverted.NewSimpleTokenizer())
	turkishAnalyzer.AddTokenFilter(inverted.NewTurkishLowercaseFilter())
	turkishAnalyzer.AddTokenFilter(inverted.NewTurkishAccentFilter())
	turkishAnalyzer.AddTokenFilter(inverted.NewTurkishStemFilter())

	bookIndex.simpleHighlighter = inverted.NewSimpleHighlighter(turkishAnalyzer)
	bookIndex.spanHighlighter = inverted.NewSpanHighlighter(turkishAnalyzer)
	bookIndex.idx = inverted.NewInvertedIndexFromFile(turkishAnalyzer, false)

	bookIndex.booksMap = deserializeBooks()
	bookIndex.pagesMap = deserializePages()

	var err error
	bookIndex.payloadStore, err = OpenCdbStore()
	if err != nil {
		log.Println(err)
	}

	return bookIndex
}

func (bi *BookIndex) reIndexFiles() error {
	fileInfos, err := ioutil.ReadDir("books")
	if err != nil {
		log.Printf("failed to open 'books' directory")
		return err
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
		}
	}

	bi.idx.UpdateAvgFieldLen()
	bi.idx.BuildCategoryBitmap()

	bi.idx.MarshalIndex()

	serializeBooks(bi.booksMap)
	serializePages(bi.pagesMap)

	return nil
}

func (bi *BookIndex) buildPayloadDatabase() error {
	payloadStore, err := NewCdbStore()
	if err != nil {
		log.Println("Failed to create cdb file")
		return err
	}

	payloadStore.BuildDatabase()
	err = payloadStore.Freeze()
	if err != nil {
		log.Println("Failed to freze payload.cdb file")
		return err
	}
	err = payloadStore.writer.Close()
	if err != nil {
		log.Println("Failed to close payload.cdb file")
		return err
	}

	payloadStore, err = OpenCdbStore()
	if err != nil {
		log.Println("Failed to open cdb file")
		return err
	}

	bi.payloadStore = payloadStore

	return nil
}

func (bi *BookIndex) GetBook(hash string) Book {
	for _, v := range bi.booksMap {
		if v.Hash == hash {
			return v
		}
	}

	return Book{}
}

func (bi *BookIndex) indexFiles(path string) {

	books, err := prepareBooks(path)
	if err != nil {
		fmt.Println("failed to load book list csv file", err)
		return
	}

	for i, book := range books {
		book.Id = uint32(i)
		bi.booksMap[book.Id] = book
		indexBook(book)
	}

	bi.idx.UpdateAvgFieldLen()
	bi.idx.BuildCategoryBitmap()

	bi.idx.MarshalIndex()

	serializeBooks(bi.booksMap)
	serializePages(bi.pagesMap)
}
