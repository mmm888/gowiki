package main

//http://mjhd.hatenablog.com/entry/my-wikisystem-made-with-golang

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/russross/blackfriday"
)

var (
	cd   string
	path []string
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
		dirname := r.FormValue("dir")
		for _, dir := range path {
			if dir == dirname {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}
		err := os.Mkdir(cd+dirname, 0755)
		if err != nil {
			fmt.Println(err)
		}
		path = append(path, dirname)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	t := template.Must(template.ParseFiles("templates/index.tmpl", "templates/base_top.tmpl"))
	err := t.Execute(w, path)
	if err != nil {
		panic(err)
	}
}

func dirHandler(w http.ResponseWriter, r *http.Request) {
	var filepath []string
	dirname := mux.Vars(r)["dir"]
	nowdir := cd + dirname
	files, err := ioutil.ReadDir(nowdir)
	if err != nil {
		fmt.Println(err)
	}
	for _, file := range files {
		filepath = append(filepath, file.Name())
	}
	t := template.Must(template.ParseFiles("templates/dirshow.tmpl", "templates/base_top.tmpl"))
	err = t.Execute(w, filepath)
	if err != nil {
		panic(err)
	}
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	dirname := mux.Vars(r)["dir"]
	filename := mux.Vars(r)["file"]
	nowpath := cd + dirname + "/" + filename
	file, err := ioutil.ReadFile(nowpath)
	if err != nil {
		fmt.Println(err)
	}
	w.Header().Add("Content-Type", "text/html")
	file_md := blackfriday.MarkdownCommon(file)
	fmt.Fprintln(w, string(file_md))
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
	cd = "data/"
	files, err := ioutil.ReadDir(cd)
	if err != nil {
		fmt.Println(err)
	}
	for _, file := range files {
		path = append(path, file.Name())
	}

	r := mux.NewRouter()
	r.HandleFunc("/", RootHandler)
	r.HandleFunc("/upload", uploadHandler)
	r.HandleFunc("/save", saveHandler)
	r.HandleFunc("/errorPage", errorPageHandler)
	r.HandleFunc("/{dir}", dirHandler)
	r.HandleFunc("/{dir}/{file}", fileHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
