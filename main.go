package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/juruen/rmapi/api.v2"
	"github.com/juruen/rmapi/auth"
)

const codeEnv string = "RMUPLOADER_CODE"

type server struct {
	cli *api.Client
}

// newServer creates a server with an api client with correct authentication
// initiated using the code environment variable.
func newServer() (server, error) {
	s := server{}

	code, ok := os.LookupEnv(codeEnv)
	if !ok {
		return s, fmt.Errorf("%s variable is not defined", codeEnv)
	}

	auth := auth.New()
	auth.RegisterDevice(code)

	// The default auth uses ~/.rmapi to store credentials
	s.cli = api.NewClient(auth.Client())

	return s, nil
}

func main() {
	s, err := newServer()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	http.HandleFunc("/", s.index)
	http.HandleFunc("/upload", s.upload)
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("web/img/"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("web/css/"))))

	if err := http.ListenAndServe(":8080", logRequest(http.DefaultServeMux)); err != nil {
		panic(err)
	}
}

func (s server) index(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("web/index.html"))
	tmpl.Execute(w, nil)
}

func (s server) upload(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		file, header, err := r.FormFile("file")
		if err != nil {
			fmt.Fprintf(w, "%v", err)
			return
		}
		defer file.Close()

		id := uuid.New().String()
		if err = s.uploadToRm(id, file, header.Filename); err != nil {
			fmt.Fprintf(w, "%v", err)
			return
		}

		fmt.Fprintf(w, id)

	case "DELETE":
		id, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, "%v", err)
			return
		}

		if err := s.deleteFromRm(string(id)); err != nil {
			fmt.Fprintf(w, "%v", err)
			return
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
