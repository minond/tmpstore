package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const LOCAL_STORE_LOC = "./uploads/"
const CHARS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const NAME_LEN = 32

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	rand.Seed(time.Now().UnixNano())

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
	name := generateRandomName()
	err = uploadLocal(name, tmpFile)
	ttl := 5 * time.Minute

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go delete(name, ttl, deleteLocal)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(fmt.Sprintf(`{"name":"%s"}`, name)))
}

func delete(name string, timeout time.Duration, handler func(string) error) {
	fmt.Printf("Deleting %s in %s\n", name, timeout.String())

	select {
	case <-time.After(timeout):
		fmt.Printf("Deleting %s now\n", name)
		err := handler(name)

		if err != nil {
			fmt.Printf("Having a reaalll hard time deleting %s\n", name)
		} else {
			fmt.Printf("Successfully deleted %s\n", name)
		}
	}
}

func deleteLocal(name string) error {
	path := LOCAL_STORE_LOC + name
	return os.Remove(path)
}

func uploadLocal(name string, f multipart.File) error {
	path := LOCAL_STORE_LOC + name
	localFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		msg := fmt.Sprintf("Could not access internal storage: %s", err)
		return errors.New(msg)
	}

	defer localFile.Close()
	io.Copy(localFile, f)

	fmt.Printf("Storing %s locally\n", name)

	return nil
}

func generateRandomName() string {
	buff := make([]byte, NAME_LEN)

	for i := range buff {
		buff[i] = CHARS[rand.Intn(len(CHARS))]
	}

	return string(buff)
}
