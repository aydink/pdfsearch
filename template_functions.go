package main

import (
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

func init() {
	funcMap = template.FuncMap{
		"inc":    Increment,
		"add":    Add,
		"tohtml": ToHtml,
		"join":   JoinStringSlice,
	}
}
