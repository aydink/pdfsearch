package main

import (
	"fmt"
	"html/template"
	"strings"
)

var funcMap template.FuncMap

func Increment(i int) int {
	return i + 1
}

func Add(a, b int) int {
	return a + b
}

func ToHtml(s string) template.HTML {
	return template.HTML(s)
}

func JoinStringSlice(s []string, separator string) string {
	return strings.Join(s, separator)
}

func Pager(numPages int, pages []PaginationItem) template.HTML {

	html := "<div class=\"pagination\">\n"

	if numPages > 10 {
		html += fmt.Sprintf("<a href=\"#\" onclick=\"gotoPage(0)\">İlk</a>\n")
	}

	for _, page := range pages {
		if page.Active {
			html += fmt.Sprintf("<a href=\"#\" class=\"active\" onclick=\"gotoPage(%d)\">%d</a>\n", page.Start, page.Page)
		} else {
			html += fmt.Sprintf("<a href=\"#\" onclick=\"gotoPage(%d)\">%d</a>\n", page.Start, page.Page)
		}
	}

	if numPages > 10 {
		html += fmt.Sprintf("<a href=\"#\" onclick=\"gotoPage(%d)\">Son</a>\n", (numPages-1)*10)
	}

	html += "</div>"

	return template.HTML(html)
}

func init() {
	funcMap = template.FuncMap{
		"inc":      Increment,
		"add":      Add,
		"tohtml":   ToHtml,
		"join":     JoinStringSlice,
		"paginate": Pager,
	}
}
