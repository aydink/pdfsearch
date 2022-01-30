package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))

	if r.Method == "GET" {

		data := make(map[string]interface{})
		t.ExecuteTemplate(w, "upload", data)

	} else {

		errorMap, err := processUploadedPdf(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			for key, val := range errorMap {
				fmt.Fprintf(w, "%s: %s\n", key, val)
			}
			return
		}
		data := make(map[string]interface{})
		data["message"] = "Dosya başarıyla yüklendi."
		t.ExecuteTemplate(w, "upload", data)
	}
}

func processUploadedPdf(r *http.Request) (map[string]string, error) {

	// max file size is 200 mb --> 209715200 bytes
	r.ParseMultipartForm(209715200)

	formErrors := make(map[string]string)
	title := strings.TrimSpace(r.PostFormValue("title"))
	if len(title) < 2 {
		formErrors["title"] = "Kitap adı 2 karakterden kısa olamaz!"
	}

	category := r.PostForm["category"]

	book := Book{}
	book.Title = title
	book.Category = category

	if len(formErrors) > 0 {
		log.Printf("/api/addbook errors:%s\n", formErrors)
		return formErrors, errors.New("uploaded form has errors")
	}

	file, _, err := r.FormFile("file")

	if err != nil {
		log.Println("FormFile:", err)
		return formErrors, err
	}
	defer file.Close()

	// Create a buffer to store the header of the file in
	fileHeader := make([]byte, 512)

	// Copy the headers into the FileHeader buffer
	if _, err := file.Read(fileHeader); err != nil {
		return formErrors, err
	}

	// set position back to start.
	if _, err := file.Seek(0, 0); err != nil {
		return formErrors, err
	}

	contentType := http.DetectContentType(fileHeader)

	if contentType == "application/pdf" {

		tempFileName := getUuid()

		f, err := os.Create("books/" + tempFileName)
		if err != nil {
			log.Println("Open file:", err)
			return formErrors, err
		}

		h := md5.New()
		multiWriter := io.MultiWriter(f, h)

		io.Copy(multiWriter, file)

		hashInBytes := h.Sum(nil)
		//Convert the bytes to a string
		md5string := hex.EncodeToString(hashInBytes)

		// close the temp file and rename it using md5 hash of the file
		f.Close()
		err = os.Rename("books/"+tempFileName, "books/"+md5string+".pdf")
		if err != nil {
			log.Println("File rename failed:", err)
			return formErrors, err
		}

		book.Hash = md5string
		//fmt.Printf("%+v\n", book)
		err = processPdfFile(book)
		if err != nil {
			return nil, err
		}

		saveBookMeta(book)
		if err != nil {
			return nil, err
		}

		indexBook(book)
	} else {
		formErrors["content_type"] = "Yalnızca PDF dosyaları desteklenmektedir. Geçerli bir dosya yükleyin."
		log.Printf("Content-Type not supported, expecting application/pdf found %s\n", contentType)
		return formErrors, fmt.Errorf("Content-Type: %s not supported, expecting application/pdf", contentType)
	}

	return formErrors, nil
}
