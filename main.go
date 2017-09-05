package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gorilla/mux"
	"github.com/russross/blackfriday"
	"github.com/unrolled/render"
)

var (
	re       *render.Render
	reponame = "wikitest/"
	subdir   = "repo/"
	dirtree  string
	//baseurl  = "http://dev01-xenial:8080"
	baseurl = "http://localhost:8080"
	actEdit = "?action=E"
	actSave = "?action=S"

	// only use RootHandler
	cd   string
	path []string
)

// RootHandler is routing of "/"
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

// Initialize is routing of "/init"
func Initialize(w http.ResponseWriter, r *http.Request) {
	err := re.HTML(w, http.StatusOK, "initialize", nil)
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusFound)
	}
}

// Settings is routing of "/setting"
func Settings(w http.ResponseWriter, r *http.Request) {
	rname := r.FormValue("rname")
	_, err := exec.Command("git", "clone", rname).Output()
	if err != nil {
		//	http.Redirect(w, r, "/error", http.StatusFound)
		log.Printf("Could not clone %s: %s", rname, err.Error())
	}
	reponame = rname
	http.Redirect(w, r, "/repo", http.StatusFound)
}

func dirHandler(w http.ResponseWriter, r *http.Request, repo Repo) {
	switch repo.act {

	// Edit Display
	case "E":
		f, err := ioutil.ReadFile(repo.rp + "/README.md")
		if err != nil {
			log.Println(err, "Cannot read file")
		}

		err = re.HTML(w, http.StatusOK, "edit_dir", struct {
			Content string
			Path    string
			Epath   string
			Spath   string
		}{
			string(f), repo.vp, repo.evp, repo.svp,
		})
		if err != nil {
			log.Println(err, "Cannot generate template")
		}

	// Save Display
	case "S":
		s := r.FormValue("submit")
		if s == "Update" {
			f, err := os.Create(repo.rp + "/README.md")
			if err != nil {
				log.Println(err, "Cannot create file")
			}
			defer f.Close()

			con := r.FormValue("content")
			_, err = f.Write([]byte(con))
			if err != nil {
				log.Println(err, "Cannot writer file")
			}

			name := r.FormValue("FileName")
			ForD := r.FormValue("ForD")
			if ForD == "File" {
				_, err = os.OpenFile(repo.rp+"/"+name, os.O_CREATE, 0644)
				if err != nil {
					log.Println(err, "Cannot create file")
				}
			} else if ForD == "Dir" {
				err = os.Mkdir(repo.rp+"/"+name, 0755)
				if err != nil {
					log.Println(err, "Cannot create directory")
				}
			}

			dirtree = ""
			dirTree(&dirtree, reponame)
		}
		http.Redirect(w, r, repo.vp, http.StatusFound)

	// Show Display
	default:
		var err error
		_, err = os.Stat(repo.rp + "/README.md")
		if err != nil {
			err = ioutil.WriteFile(repo.rp+"/README.md", nil, 0644)
			if err != nil {
				log.Println(err, "Cannot create README.md")
			}
		}
		f, err := ioutil.ReadFile(repo.rp + "/README.md")
		if err != nil {
			log.Println(err, "Cannot read file")
		}

		// redirect "edit" when content is ""
		if string(f) == "" {
			http.Redirect(w, r, repo.evp, http.StatusFound)
		}

		md := blackfriday.MarkdownCommon(f)
		err = re.HTML(w, http.StatusOK, "repo", struct {
			Content string
			Path    string
			Epath   string
			Spath   string
			Dirtree string
		}{
			string(md), repo.vp, repo.evp, repo.svp, dirtree,
		})
		if err != nil {
			log.Println(err, "Cannot generate template")
		}
	}
}

func fileHandler(w http.ResponseWriter, r *http.Request, repo Repo) {
	switch repo.act {

	// Edit Display
	case "E":
		f, err := ioutil.ReadFile(repo.rp)
		if err != nil {
			log.Println(err, "Cannot read file")
		}

		err = re.HTML(w, http.StatusOK, "edit_file", struct {
			Content  string
			Path     string
			Epath    string
			Spath    string
			FileName string
		}{
			string(f), repo.vp, repo.evp, repo.svp, "TEST",
		})
		if err != nil {
			log.Println(err, "Cannot generate template")
		}

	// Save Display
	case "S":
		s := r.FormValue("submit")
		if s == "Save" {
			f, err := os.Create(repo.rp)
			if err != nil {
				log.Println(err, "Cannot create file")
			}
			defer f.Close()

			con := r.FormValue("content")
			_, err = f.Write([]byte(con))
			if err != nil {
				log.Println(err, "Cannot writer file")
			}
		}
		http.Redirect(w, r, repo.vp, http.StatusFound)

	// Show Display
	default:
		f, err := ioutil.ReadFile(repo.rp)
		if err != nil {
			log.Println(err, "Cannot read file")
		}

		// redirect "edit" when content is ""
		if string(f) == "" {
			http.Redirect(w, r, repo.evp, http.StatusFound)
		}

		md := blackfriday.MarkdownCommon(f)
		err = re.HTML(w, http.StatusOK, "repo", struct {
			Content string
			Path    string
			Epath   string
			Spath   string
			Dirtree string
		}{
			string(md), repo.vp, repo.evp, repo.svp, dirtree,
		})
		if err != nil {
			log.Println(err, "Cannot generate template")
		}
	}
}

func repoHandler(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	var repo Repo
	repo.act = r.FormValue("action")
	repo.rp = reponame + path
	repo.vp = r.URL.String()
	if strings.HasSuffix(repo.vp, actEdit) {
		repo.vp = strings.TrimSuffix(repo.vp, actEdit)
	}
	repo.evp = repo.vp + actEdit
	if strings.HasSuffix(repo.vp, actSave) {
		repo.vp = strings.TrimSuffix(repo.vp, actSave)
	}
	repo.svp = repo.vp + actSave

	// check whether file or directory
	f, err := os.Stat(repo.rp)
	if err != nil {
		log.Println(err, "Failure to checking if file exists")
	}
	if f.IsDir() {
		dirHandler(w, r, repo)
	} else {
		fileHandler(w, r, repo)
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

// ErrorPage is routing of "/error"
func ErrorPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", "<p>Internal Server Error</p>")
}

// TestHandler is routing of "/test"
func TestHandler(w http.ResponseWriter, r *http.Request) {
	//tmpl := template.Must(template.ParseFiles("./test.html"))
	tmpl := template.Must(template.ParseFiles("./test2.html"))
	tmpl.Execute(w, nil)
}

func main() {
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

	r.HandleFunc("/repo", repoHandler)
	p := r.PathPrefix("/repo").Subrouter()
	p.HandleFunc("/{path:.*}", repoHandler)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	log.Fatal(http.ListenAndServe(":8080", r))
}
