package main

type Book struct {
	Id       uint32   `json:"id"`
	Title    string   `json:"title"`
	Category []string `json:"category"`
	NumPages uint32   `json:"num_pages"`
	Hash     string   `json:"hash"`
}
