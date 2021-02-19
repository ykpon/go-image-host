package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// Result struct for output json
type Result struct {
	Path string
	Name string
	Host string
}

const publicDir = "public"
const uploadDir = "/uploads"

var filenameLength *int
var hostname *string

// Create directory
func createDirectory(path string) {
	err := os.MkdirAll(publicDir+path, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(publicDir + path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Letters for random string generator
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// File size in MB
const fileSize = 10

func uploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(fileSize << 20)
	file, _, err := r.FormFile("image")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()

	year, month, day := time.Now().Date()
	dirName := fmt.Sprintf(uploadDir+"/%d/%d/%d", year, month, day)
	createDirectory(dirName)
	var filename string

	for i := 0; i < 5; i++ {
		tempFilename := randSeq(*filenameLength)
		if fileExists(dirName + "/" + tempFilename) {
			continue
		}
		filename = tempFilename
	}

	if filename == "" {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filepath := fmt.Sprintf(dirName+"/%s.png", filename)
	dst, err := os.Create(publicDir + filepath)
	defer dst.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := Result{Path: filepath, Name: filename, Host: *hostname}
	js, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"]", "[IP: "+r.RemoteAddr+"]", "Uploading new file: \""+filepath+"\"")
}

func setupRoutes(frontendHandle bool) {
	if frontendHandle {
		http.Handle("/", http.FileServer(http.Dir("./public")))
		fmt.Println("Frontend handle added.")
	}

	http.HandleFunc("/upload", uploadFile)
	http.ListenAndServe(":8080", nil)
}

func main() {
	var frontendHandle = flag.Bool("frontend", false, "Add handle for / static files")
	filenameLength = flag.Int("filename-length", 6, "Set max length for filename")
	hostname = flag.String("hostname", "localhost", "Hostname of your server")
	flag.Parse()
	fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"]", "Server started...")
	setupRoutes(*frontendHandle)
}
