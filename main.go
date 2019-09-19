package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/PuerkitoBio/goquery"
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

	log.Println("Setting up authentication with the device...")
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

	// TODO(lobre): GET should return the index page with the msg GET parameters processed
	// TODO(lobre): POST should fetch from the url and upload. An error can directly be provided in the answer
	http.HandleFunc("/", s.index)
	// TODO(lobre): Ajax request to upload a file. In javascript, we should handle the error and redirect to / with msg.
	// Another solution to avoid a redirect could be to directly change the error toast from javascript.
	// That could allow to remove the GET parameter msg behavior from the / index page.
	http.HandleFunc("/upload", s.upload)
	// TODO(lobre): Ajax request to revert a file. In javascript, we should handle the error and redirect to / with msg.
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
	msg := r.FormValue("msg")
	tmpl := template.Must(template.ParseFiles("web/index.html"))
	tmpl.Execute(w, msg)
}

func (s server) upload(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Defines a new UUID for the document to be uploaded.
		id := uuid.New().String()

		if r.FormValue("file") == "" && r.FormValue("url") == "" {
			http.Redirect(w, r, queryMsg("invalid form sent"), http.StatusFound)
			return
		}

		// if url is defined, we try to generate a PDF
		// from the targeted website.
		// Otherwise, we push the provided file.
		url := r.FormValue("url")
		if url != "" {
			var err error
			file, err := webpageAsPDF(url)
			if err != nil {
				http.Redirect(w, r, queryMsg(err.Error()), http.StatusFound)
				return
			}

			name, err := titleFromURL(url)
			if err != nil {
				http.Redirect(w, r, queryMsg(err.Error()), http.StatusFound)
				return
			}

			if err := s.uploadToRm(id, file, name); err != nil {
				http.Redirect(w, r, queryMsg(err.Error()), http.StatusFound)
				return
			}

			http.Redirect(w, r, queryMsg("document has been sent"), http.StatusFound)
			return

		} else {

			fr, header, err := r.FormFile("file")
			if err != nil {
				http.Redirect(w, r, queryMsg(err.Error()), http.StatusFound)
				return
			}
			defer fr.Close()

			file, err := ioutil.ReadAll(fr)
			if err != nil {
				http.Redirect(w, r, queryMsg(err.Error()), http.StatusFound)
				return
			}

			if err := s.uploadToRm(id, file, header.Filename); err != nil {
				http.Redirect(w, r, queryMsg(err.Error()), http.StatusFound)
				return
			}

			// send id as result
			fmt.Fprintf(w, id)
		}

	case "DELETE":
		id, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Redirect(w, r, queryMsg(err.Error()), http.StatusFound)
			return
		}

		if err := s.deleteFromRm(string(id)); err != nil {
			http.Redirect(w, r, queryMsg(err.Error()), http.StatusFound)
			return
		}

	default:
		http.Redirect(w, r, queryMsg("http method not supported"), http.StatusFound)
		return
	}

}

func (s server) delete(w http.ResponseWriter, r *http.Request) {
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

// queryMsg returns a formatted url query string with
// the msg parameter defined and encoded.
func queryMsg(msg string) string {
	return fmt.Sprintf("?msg=%s", url.QueryEscape(msg))
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
