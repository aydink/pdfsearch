package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func getUuid() (uuid string) {

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Println("Error: ", err)
		return
	}

	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	//fmt.Println(uuid)

	return
}

func preparePdfFile(path string) (string, error) {

	var md5string string

	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Println("Path:", path, err)
		return md5string, err
	}

	// Create a buffer to store the header of the file in
	fileHeader := make([]byte, 512)

	// Copy the headers into the FileHeader buffer
	if _, err := file.Read(fileHeader); err != nil {
		return md5string, err
	}

	// set position back to start.
	if _, err := file.Seek(0, 0); err != nil {
		return md5string, err
	}

	contentType := http.DetectContentType(fileHeader)

	if contentType == "application/pdf" {

		tempFileName := getUuid()

		f, err := os.Create("books/" + tempFileName)
		if err != nil {
			log.Println("OpenFile:", err)
			return md5string, err
		}

		h := md5.New()
		multiWriter := io.MultiWriter(f, h)

		io.Copy(multiWriter, file)

		hashInBytes := h.Sum(nil)
		//Convert the bytes to a string
		md5string = hex.EncodeToString(hashInBytes)

		// close the temp file and rename it using md5 hash of the file
		f.Close()
		err = os.Rename("books/"+tempFileName, "books/"+md5string+".pdf")
		if err != nil {
			log.Println("File rename failed:", err)
			return md5string, err
		}

	} else {
		log.Printf("Content-Type not supported, expecting application/pdf found %s\n", contentType)
		return md5string, fmt.Errorf("Content-Type %s not supported, expecting application/pdf", contentType)
	}

	return md5string, nil
}

func processPdfFile(book Book) error {

	output, err := exec.Command("pdfinfo", "books/"+book.Hash+".pdf").Output()
	if err != nil {
		log.Println(err)
		return err
	}

	re := regexp.MustCompile("Pages: *([0-9]+)")
	matches := re.FindStringSubmatch(string(output))
	if len(matches) == 2 {
		numPages, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Printf("Failed to find PDF file number of pages, file:%s.pdf, error:%s\n", err, book.Hash)
			return err
		}

		book.NumPages = uint32(numPages)
	}

	if _, err := os.Stat("books/" + book.Hash + ".txt"); os.IsNotExist(err) {
		_, err = exec.Command("pdftotext", "-enc", "UTF-8", "books/"+book.Hash+".pdf", "books/"+book.Hash+".txt").Output()
		if err != nil {
			//log.Fatalln(err)
			log.Printf("PDF text extraction failed, file:%s.pdf, error:%s\n", err, book.Hash)
			return err
		}
	}

	if _, err := os.Stat("books/" + book.Hash + ".bbox.txt"); os.IsNotExist(err) {
		_, err = exec.Command("pdftotext", "-enc", "UTF-8", "-bbox", "books/"+book.Hash+".pdf", "books/"+book.Hash+".bbox.txt").Output()
		if err != nil {
			//log.Fatalln(err)
			log.Printf("PDF payload extraction failed, file:%s.pdf, error:%s\n", err, book.Hash)
			return err
		}
	}

	if _, err := os.Stat("books/" + book.Hash + ".bbox.gob"); os.IsNotExist(err) {
		f, err := os.Create("books/" + book.Hash + ".bbox.gob")
		defer f.Close()
		if err != nil {
			log.Println(err)
			return err
		}

		log.Println("-------------------------------", book.Hash)
		payload, err := CreatePayload(book.Hash)
		if err != nil {
			log.Println(err)
			return err
		}

		_, err = f.Write(payload)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func parseTextFile(hash string) ([]string, error) {

	content, err := ioutil.ReadFile("books/" + hash + ".txt")

	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	pages := strings.Split(string(content), "\f")
	return pages, nil
}

func getPages(book Book) ([]Page, error) {

	pages, err := parseTextFile(book.Hash)
	if err != nil {
		log.Printf("Parsing pdf text file failed, file:%s.pdf, error:%s\n", err, book.Hash)
		return nil, err
	}

	numPages := len(pages)

	docs := make([]Page, 0)

	for i := 0; i < numPages; i++ {

		doc := Page{}
		doc.BookId = book.Id
		doc.Content = pages[i]
		doc.PageNumber = uint32(i + 1)

		docs = append(docs, doc)
	}

	return docs, nil
}
