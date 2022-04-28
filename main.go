package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

var flagReindex *bool
var flagPath *string
var flagPort *int
var flagInMemory *bool
var flagEnablePublicNetwork *bool

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
	flagInMemory = flag.Bool("inmemory", false, "create an inmemory index or open from disk")
	flagEnablePublicNetwork = flag.Bool("public", false, "enable access from local network")
	flagPath = flag.String("path", "pdf", "path to pdf files that will be indexed")
	flagPort = flag.Int("port", 8080, "http server port, default to 8080")

	flag.Parse()

	if (*flagPort < 1) || (*flagPort > 65535) {
		log.Fatalln("http port is not valid, must be within 1 to 65535 range")
	}

	port := strconv.Itoa(*flagPort)

	fmt.Printf("reindex index: %t\n", *flagReindex)
	fmt.Printf("inmemory index:%t\n", *flagInMemory)
	fmt.Printf("enable network access:%t\n", *flagEnablePublicNetwork)

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
	fmt.Println("Arama motorunu kullanmak için tarayıcı ile http://127.0.0.1:" + port + " adresine gidin")

	host := "127.0.0.1:" + port

	if *flagEnablePublicNetwork {
		host = ":" + port
	}

	openBrowser("http://127.0.0.1:" + port)

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
