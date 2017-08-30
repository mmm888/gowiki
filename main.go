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
	"os/exec"

	"github.com/gorilla/mux"
	"github.com/russross/blackfriday"
	"github.com/unrolled/render"
)

var (
	re   *render.Render
	repo string
	cd   string
	path []string
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	if action == "EDIT" {
		err := re.HTML(w, http.StatusOK, "edit_dir", nil)
		if err != nil {
			http.Redirect(w, r, "/error", http.StatusFound)
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
			log.Printf("error: %s", err.Error())
		}
		path = append(path, dirname)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	err := re.HTML(w, http.StatusOK, "index", path)
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
	}
}

func Initialize(w http.ResponseWriter, r *http.Request) {
	err := re.HTML(w, http.StatusOK, "initialize", nil)
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
	}
}

func Settings(w http.ResponseWriter, r *http.Request) {
	rname := r.FormValue("rname")
	_, err := exec.Command("git", "clone", rname).Output()
	if err != nil {
		//	http.Redirect(w, r, "/error", http.StatusFound)
		log.Printf("Could not clone %s: %s", rname, err.Error())
	}
	repo = rname
	http.Redirect(w, r, "/repo", http.StatusFound)
}

func Repository(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	rp := repo + path
	vp := r.URL.String() + "/"
	f, err := os.Stat(rp)
	if err != nil {
		log.Println("error: %s", err.Error())
	}
	if f.IsDir() {
		var files []string
		var err error
		dir, err := ioutil.ReadDir(rp)
		if err != nil {
			log.Printf("error: %s", err.Error())
		}
		for _, f := range dir {
			files = append(files, f.Name())
		}
		err = re.HTML(w, http.StatusOK, "repo_dir", struct {
			Files []string
			Path  string
		}{
			files, vp,
		})
	} else {
		file, err := ioutil.ReadFile(rp)
		if err != nil {
			log.Printf("error: %s", err.Error())
		}
		file_md := blackfriday.MarkdownCommon(file)
		err = re.HTML(w, http.StatusOK, "repo_file", struct {
			Content string
		}{
			string(file_md),
		})
		if err != nil {
			http.Redirect(w, r, "/error", http.StatusFound)
		}
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
			http.Redirect(w, r, "/error", http.StatusFound)
		}

		_, err = io.Copy(uploadedFile, part)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			uploadedFile.Close()
			http.Redirect(w, r, "/error", http.StatusFound)
		}
	}
	http.Redirect(w, r, "/upload", http.StatusFound)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	err := re.HTML(w, http.StatusOK, "upload", nil)
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
	}
}

func ErrorPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", "<p>Internal Server Error</p>")
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./test.html"))
	tmpl.Execute(w, nil)
}

func main() {
	baseurl := "http://dev01-xenial:8080"
	repo = "wikitest/"

	re = render.New(render.Options{
		Directory: "templates",
		Funcs: []template.FuncMap{
			{
				"url_for":  func(path string) string { return baseurl + path },
				"safehtml": func(text string) template.HTML { return template.HTML(text) },
				"stradd":   func(a string, b string) string { return a + b },
			},
		},
	})

	r := mux.NewRouter()
	r.HandleFunc("/", RootHandler)
	r.HandleFunc("/init", Initialize)
	r.HandleFunc("/setting", Settings)
	r.HandleFunc("/upload", uploadHandler)
	r.HandleFunc("/save", saveHandler)
	r.HandleFunc("/error", ErrorPage)
	r.HandleFunc("/test", TestHandler)

	r.HandleFunc("/repo", Repository)
	p := r.PathPrefix("/repo/").Subrouter()
	p.HandleFunc("/{path:.*}", Repository)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	log.Fatal(http.ListenAndServe(":8080", r))
}
