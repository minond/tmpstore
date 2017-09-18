package main

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/ping", ping)
	http.HandleFunc("/upload", upload)
	http.ListenAndServe(":"+port, nil)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "plain/text")
	w.Write([]byte("pong"))
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Unsupported method", http.StatusNotFound)
		return
	}

	r.ParseMultipartForm(32 << 20)
	tmpFile, handler, err := r.FormFile("file")

	if err != nil {
		msg := fmt.Sprintf("Could not extract file from request: %s", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	defer tmpFile.Close()
	fmt.Fprintf(w, "%v", handler.Header)

	err = uploadLocal(handler.Filename, tmpFile)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func uploadLocal(name string, f multipart.File) error {
	localFile, err := os.OpenFile("./uploads/"+name, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		msg := fmt.Sprintf("Could not access internal storage: %s", err)
		return errors.New(msg)
	}

	defer localFile.Close()
	io.Copy(localFile, f)

	return nil
}
