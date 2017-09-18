package main

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
)

var LOCAL_STORE_LOC = "./uploads/"

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/ping", ping)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/get/", serve)
	http.ListenAndServe(":"+port, nil)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "plain/text")
	w.Write([]byte("pong"))
}

func serve(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[5:]
	http.ServeFile(w, r, LOCAL_STORE_LOC+name)
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Unsupported method", http.StatusNotFound)
		return
	}

	r.ParseMultipartForm(32 << 20)
	tmpFile, _, err := r.FormFile("file")

	if err != nil {
		msg := fmt.Sprintf("Could not extract file from request: %s", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	defer tmpFile.Close()
	name, err := generateRandomName()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = uploadLocal(name, tmpFile)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "plain/text")
	w.Write([]byte(name))
}

func uploadLocal(name string, f multipart.File) error {
	localFile, err := os.OpenFile(LOCAL_STORE_LOC+name, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		msg := fmt.Sprintf("Could not access internal storage: %s", err)
		return errors.New(msg)
	}

	defer localFile.Close()
	io.Copy(localFile, f)

	return nil
}

func generateRandomName() (string, error) {
	out, err := exec.Command("uuidgen").Output()

	if err != nil {
		return "", err
	}

	str := string(out[:])
	str = str[:len(str)-5]

	return str, nil
}
