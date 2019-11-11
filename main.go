package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/juruen/rmapi/auth"
	"github.com/juruen/rmapi/cloud"
)

const codeEnv string = "RMUPLOADER_CODE"

type server struct {
	cli *cloud.Client
}

// newServer creates a server with an api client with correct authentication
// initiated using the code environment variable.
func newServer() (server, error) {
	s := server{}

	code, ok := os.LookupEnv(codeEnv)
	if !ok {
		return s, fmt.Errorf("%s variable is not defined", codeEnv)
	}

	log.Println("Setting up authentication with the device...")
	auth := auth.New()
	auth.RegisterDevice(code)

	// The default auth uses ~/.rmapi to store credentials
	s.cli = cloud.NewClient(auth.Client())

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
	http.HandleFunc("/delete", s.delete)

	// Statics
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("web/img/"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("web/css/"))))

	log.Println("Starting web server...")
	if err := http.ListenAndServe(":8080", logRequest(http.DefaultServeMux)); err != nil {
		panic(err)
	}
}

func (s server) index(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("web/index.html"))
	msg := ""

	switch r.Method {
	case "GET":
	case "POST":
		url := r.FormValue("url")
		if url == "" {
			msg = "url not provided"
			break
		}

		file, err := webpageAsPDF(url)
		if err != nil {
			log.Fatal(err)
			msg = err.Error()
			break
		}

		name, err := titleFromURL(url)
		if err != nil {
			log.Fatal(err)
			msg = err.Error()
			break
		}

		id := uuid.New().String()
		if err := s.uploadToRm(id, file, name); err != nil {
			log.Fatal(err)
			msg = err.Error()
			break
		}

		w.WriteHeader(http.StatusOK)
		msg = "webpage has been sent"

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		msg = http.StatusText(http.StatusMethodNotAllowed)
	}

	tmpl.Execute(w, msg)
}

func (s server) upload(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	// called using ajax
	case "POST":
		if r.FormValue("file") == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "missing file parameter")
			return
		}

		f, header, err := r.FormFile("file")
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, err.Error())
			return
		}
		defer f.Close()

		file, err := ioutil.ReadAll(f)
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, err.Error())
			return
		}

		id := uuid.New().String()
		if err := s.uploadToRm(id, file, header.Filename); err != nil {
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

		// send id as result
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, id)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

}

func (s server) delete(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	// called using ajax
	case "DELETE":
		id, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

		if err := s.deleteFromRm(string(id)); err != nil {
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

		// send id as result
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, string(id))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, http.StatusText(http.StatusMethodNotAllowed))
		return
	}
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

// titleFromURL parses an URL to return a
// more friendly name.
// The extension .pdf is added at the end.
func titleFromURL(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	title := doc.Find("title").Text()

	return fmt.Sprintf("%s.pdf", title), nil
}
