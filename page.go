package main

import (
	"bytes"
	"log"

	"github.com/colinmarc/cdb"
)

type Page struct {
	Id         uint32 `json:"id"`
	BookId     uint32 `json:"book_id"`
	Content    string `json:"content"`
	PageNumber uint32 `json:"page_number"`
}

func serializePages(pages map[uint32]Page) {

	writer, err := cdb.Create("data/pages.cdb")
	if err != nil {
		log.Fatal(err)
	}

	for key, page := range pages {
		var buf bytes.Buffer
		buf.Write(uint32ToBytes(page.Id))
		buf.Write(uint32ToBytes(page.BookId))
		buf.Write(uint32ToBytes(page.PageNumber))
		buf.WriteString(page.Content)

		writer.Put(uint32ToBytes(key), buf.Bytes())
	}

	writer.Freeze()
	writer.Close()
}

func deserializePages() map[uint32]Page {

	pages := make(map[uint32]Page)

	reader, err := cdb.Open("data/pages.cdb")
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	iter := reader.Iter()
	for iter.Next() {
		buf := iter.Value()

		page := Page{}
		page.Id = bytesToUint32le(buf[0:4])
		page.BookId = bytesToUint32le(buf[4:8])
		page.PageNumber = bytesToUint32le(buf[8:12])
		page.Content = string(buf[12:])

		pages[page.Id] = page
	}

	return pages
}
