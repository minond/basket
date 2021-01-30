package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

const (
	formFileName = "data"
)

var (
	listen        = flag.String("listen", ":8080", "host and port to bind http server to")
	maxUploadSize = flag.Int64("max-upload-size", 10<<20, "maximum file size")
	uploadDir     = flag.String("upload-dir", "uploads", "path to local uploads directory")
)

func setup() {
	if _, err := os.Stat(*uploadDir); os.IsNotExist(err) {
		if err := os.Mkdir(*uploadDir, os.ModePerm); err != nil {
			log.Fatalf("unable to create upload directory: %s\n", err)
		}
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(*maxUploadSize)

	uploadedFile, handler, err := r.FormFile(formFileName)
	if err != nil {
		log.Printf("error reading request: %s\n", err)
		fmt.Fprintf(w, "error: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer uploadedFile.Close()

	log.Printf("storing %s (%v bytes)\n", handler.Filename, handler.Size)

	localFilePath := path.Join(*uploadDir, handler.Filename)
	localFile, err := os.OpenFile(localFilePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("error opening local file: %s\n", err)
		fmt.Fprintf(w, "error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer localFile.Close()

	// TODO don't load contents into mem
	contents, err := ioutil.ReadAll(uploadedFile)
	if err != nil {
		log.Printf("error reading contents: %s\n", err)
		fmt.Fprintf(w, "error: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	localFile.Write(contents)
}

func main() {
	flag.Parse()
	setup()
	http.HandleFunc("/", upload)
	log.Printf("listening on %s\n", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
