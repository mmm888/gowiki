package main

//http://mjhd.hatenablog.com/entry/my-wikisystem-made-with-golang

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	if action == "EDIT" {
		t := template.Must(template.ParseFiles("templates/edit_dir.tmpl", "templates/base_top.tmpl"))
		err := t.Execute(w, nil)
		if err != nil {
			panic(err)
		}
		return
	} else if action == "CREATE" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	t := template.Must(template.ParseFiles("templates/index.tmpl", "templates/base_top.tmpl"))
	err := t.Execute(w, "hello")
	if err != nil {
		panic(err)
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		if part.FileName() == "" {
			continue
		}

		uploadedFile, err := os.Create("data/" + part.FileName())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			uploadedFile.Close()
			redirectToErrorPage(w, r)
		}

		_, err = io.Copy(uploadedFile, part)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			uploadedFile.Close()
			redirectToErrorPage(w, r)
		}
	}
	http.Redirect(w, r, "/upload", http.StatusFound)
}

func showHandler(w http.ResponseWriter, r *http.Request) {
	filename := mux.Vars(r)["file"]
	fmt.Fprintf(w, "%s\n", filename)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/upload.tmpl", "templates/base_top.tmpl")
	err := t.Execute(w, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func errorPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", "<p>Internal Server Error</p>")
}

func redirectToErrorPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/errorPage", http.StatusFound)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", RootHandler)
	r.HandleFunc("/upload", uploadHandler)
	r.HandleFunc("/save", saveHandler)
	r.HandleFunc("/errorPage", errorPageHandler)

	l := r.PathPrefix("/dir/{file}")
	l.Methods("GET").HandlerFunc(showHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
