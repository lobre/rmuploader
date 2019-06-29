package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/upload", upload)
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("web/img/"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("web/css/"))))

	if err := http.ListenAndServe(":8080", logRequest(http.DefaultServeMux)); err != nil {
		panic(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("web/index.html"))
	tmpl.Execute(w, nil)
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}
	defer file.Close()

	err = saveFile(file, header.Filename)
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}

	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func saveFile(r io.Reader, name string) error {
	dirPath, err := currentPath()
	if err != nil {
		return err
	}
	dirPath = filepath.Join(dirPath, "uploads")

	// Make sure directory exists
	_ = os.Mkdir(dirPath, 0755)

	file, err := os.OpenFile(filepath.Join(dirPath, name), os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}
	defer file.Close()

	io.Copy(file, r)

	return nil
}

func currentPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}

	exPath := filepath.Dir(ex)
	return exPath, nil
}
