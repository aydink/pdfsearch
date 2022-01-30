package main

import (
	"math"
)

type PaginationItem struct {
	Page   int
	Start  int
	Active bool
}

func Paginate(start, size, numHits int) []PaginationItem {

	if start > numHits {
		start = 0
	}

	page := start / size
	//fmt.Println("page:", page)

	numPages := int(math.Ceil(float64(numHits) / float64(size)))

	pageStart := 0
	pageEnd := numPages

	if numPages > 10 {
		if (page + 5) < numHits {
			pageStart = page - 5
		}

		if pageStart < 0 {
			pageStart = 0
		}

		if (numPages - pageStart) < 10 {
			pageStart = numPages - 10
		}

		if (pageStart + 10) > numPages {
			pageEnd = numPages
		} else {
			pageEnd = pageStart + 10
		}

	}

	//fmt.Println("numPages:", numPages, "currentPage:", page, "pageStart:", pageStart, "pageEnd:", pageEnd)

	pages := make([]PaginationItem, 0, 10)

	for i := pageStart; i < pageEnd; i++ {
		p := PaginationItem{Page: i + 1, Start: i * 10, Active: i == page}
		pages = append(pages, p)
	}

	return pages
}
