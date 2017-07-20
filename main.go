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

type DirShow struct {
	Dirname    string
	Uploadpath string
	Filename   []string
}

type FileShow struct {
	Editpath string
	Content  string
}

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
	action := r.FormValue("action")
	if action == "UPLOAD" {
		t, _ := template.ParseFiles("templates/upload.tmpl", "templates/base_top.tmpl")
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	var dirshow DirShow
	dirshow.Dirname = mux.Vars(r)["dir"]
	nowdir := cd + dirshow.Dirname
	dirshow.Uploadpath = "/" + dirshow.Dirname + "?action=UPLOAD"
	files, err := ioutil.ReadDir(nowdir)
	if err != nil {
		fmt.Println(err)
	}
	for _, file := range files {
		dirshow.Filename = append(dirshow.Filename, file.Name())
	}
	t := template.Must(template.ParseFiles("templates/dirshow.tmpl", "templates/base_top.tmpl"))
	err = t.Execute(w, dirshow)
	if err != nil {
		panic(err)
	}
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	var fileshow FileShow
	dirname := mux.Vars(r)["dir"]
	filename := mux.Vars(r)["file"]
	nowpath := cd + dirname + "/" + filename
	file, err := ioutil.ReadFile(nowpath)
	if err != nil {
		fmt.Println(err)
	}
	action := r.FormValue("action")
	if action == "EDIT" {
		t, _ := template.ParseFiles("templates/edit_file.tmpl", "templates/base_top.tmpl")
		err := t.Execute(w, string(file))
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	funcMap := template.FuncMap{
		"safehtml": func(text string) template.HTML { return template.HTML(text) },
	}
	t := template.Must(template.New("fileshow.tmpl").Funcs(funcMap).ParseFiles("templates/fileshow.tmpl", "templates/base_top.tmpl"))
	file_md := blackfriday.MarkdownCommon(file)
	fileshow.Editpath = "/" + dirname + "/" + filename + "?action=EDIT"
	fileshow.Content = string(file_md)
	fmt.Println(filename)
	fmt.Println(fileshow.Editpath)
	err = t.Execute(w, fileshow)
	if err != nil {
		fmt.Println(err)
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
