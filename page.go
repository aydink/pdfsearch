package main

type Page struct {
	Id         uint32 `json:"id"`
	BookId     uint32 `json:"book_id"`
	Content    string `json:"content"`
	PageNumber uint32 `json:"page_number"`
}
