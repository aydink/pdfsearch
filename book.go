package main

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/colinmarc/cdb"
)

type Book struct {
	Id       uint32   `json:"id"`
	Title    string   `json:"title"`
	Category []string `json:"category"`
	NumPages uint32   `json:"num_pages"`
	Hash     string   `json:"hash"`
}

func serializeBooks(books map[uint32]Book) {

	writer, err := cdb.Create("data/books.cdb")
	if err != nil {
		log.Fatal(err)
	}

	for key, book := range books {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)

		if err := enc.Encode(book); err != nil {
			log.Fatal(err)
		}

		writer.Put(uint32ToBytes(key), buf.Bytes())
	}

	writer.Freeze()
	writer.Close()
}

func deserializeBooks() map[uint32]Book {

	books := make(map[uint32]Book)

	reader, err := cdb.Open("data/books.cdb")
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	iter := reader.Iter()
	for iter.Next() {
		key := bytesToUint32le(iter.Key())
		value := iter.Value()

		buf := bytes.NewBuffer(value)
		dec := gob.NewDecoder(buf)

		book := Book{}

		if err := dec.Decode(&book); err != nil {
			log.Fatal(err)
		}

		books[key] = book
	}

	return books
}
