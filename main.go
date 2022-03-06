package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var flagReindex *bool
var flagPath *string
var flagInMemory *bool
var flagEnableNetwork *bool

var bookIndex *BookIndex

func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func initBookIndex() {

	if isFlagSet("path") {

		_, err := isDirectory(*flagPath)
		if err != nil {
			log.Fatalf("provided path is not a valid")
		}

		bookIndex = NewBookIndex()
		bookIndex.indexFiles(*flagPath)
		bookIndex.buildPayloadDatabase()

		return
	}

	if *flagReindex {
		bookIndex = NewBookIndex()
		bookIndex.reIndexFiles()
		bookIndex.buildPayloadDatabase()
	} else {
		bookIndex = OpenBookIndex()
	}
}

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flagReindex = flag.Bool("reindex", false, "rebuild inmemory index from existing pdf files")
	flagInMemory = flag.Bool("inmemory", false, "create an inmemormoy index or open from disk")
	flagEnableNetwork = flag.Bool("network", false, "enable access from local network")
	flagPath = flag.String("path", "pdf", "path to pdf files that will be indexed")

	flag.Parse()

	fmt.Printf("reindex index: %t\n", *flagReindex)
	fmt.Printf("inmemory index:%t\n", *flagInMemory)
	fmt.Printf("enable network access:%t\n", *flagEnableNetwork)

	// capture Ctrl-C exit event
	cleanUpBeforeExit()

	// build fulltext index
	initBookIndex()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/test", test)
	http.HandleFunc("/search/", searchHandler)
	http.HandleFunc("/page", pageHandler)
	http.HandleFunc("/image", imageHandler)
	http.HandleFunc("/download/", downloadHandler)
	http.HandleFunc("/stats", tokenStatHandler)
	http.HandleFunc("/books", booksHandler)
	http.HandleFunc("/api/addbook", uploadHandler)
	http.HandleFunc("/api/payloads", payloadHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	fmt.Println("--------------------------------------------------")
	fmt.Println("Arama motorunu kullanmak için tarayıcı ile http://127.0.0.1:8080 adresine gidin")

	openBrowser("http://127.0.0.1:8080/")

	host := "127.0.0.1:8080"

	if *flagEnableNetwork {
		host = ":8080"
	}
	err := http.ListenAndServe(host, nil)
	if err != nil {
		fmt.Println(err)
	}

}

func cleanUpBeforeExit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			fmt.Println(sig.String(), "Ctrl-C captured")
			//payloadStore.reader.Close()
			os.Exit(0)
		}
	}()
}
