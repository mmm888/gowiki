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
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/russross/blackfriday"
	"github.com/unrolled/render"
)

var (
	re       *render.Render
	repoName = "wikitest"
	subDir   = "repo"
	dirTree  string
	protocol = "http://"
	//baseurl  = "dev01-xenial:8080"
	baseurl = "localhost:8080"
	actEdit = "?action=E"
	actSave = "?action=S"

	// only use RootHandler
	cd string
	p  []string
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
		for _, dir := range p {
			if dir == dirname {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}
		err := os.Mkdir(cd+dirname, 0755)
		if err != nil {
			log.Printf("error: %s", err.Error())
		}
		p = append(p, dirname)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	err := re.HTML(w, http.StatusOK, "index", p)
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
	repoName = rname
	http.Redirect(w, r, "/repo", http.StatusFound)
}

func initRepo(r *http.Request) Repo {
	var repo Repo
	p := mux.Vars(r)["path"]
	repo.act = r.FormValue("action")
	repo.rp = filepath.Join(repoName, p)
	repo.vp = r.URL.String()
	if strings.HasSuffix(repo.vp, actEdit) {
		repo.vp = strings.TrimSuffix(repo.vp, actEdit)
	}
	repo.evp = repo.vp + actEdit
	if strings.HasSuffix(repo.vp, actSave) {
		repo.vp = strings.TrimSuffix(repo.vp, actSave)
	}
	repo.svp = repo.vp + actSave
	return repo
}

func dirHandler(w http.ResponseWriter, r *http.Request, repo Repo) {
	switch repo.act {

	// Edit Display
	case "E":
		readmePath := filepath.Join(repo.rp, "README.md")
		f, err := ioutil.ReadFile(readmePath)
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
			Content  string
			Path     string
			Epath    string
			Spath    string
			Dirtree  string
			LinkPath string
		}{
			string(md), repo.vp, repo.evp, repo.svp, dirTree, createLinkPath(repo.vp),
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
			string(f), repo.vp, repo.evp, repo.svp, filepath.Base(repo.rp),
		})
		if err != nil {
			log.Println(err, "Cannot generate template")
		}

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
			Content  string
			Path     string
			Epath    string
			Spath    string
			Dirtree  string
			LinkPath string
		}{
			string(md), repo.vp, repo.evp, repo.svp, dirTree, createLinkPath(repo.vp),
		})
		if err != nil {
			log.Println(err, "Cannot generate template")
		}
	}
}

func repoHandler(w http.ResponseWriter, r *http.Request) {
	repo := initRepo(r)

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

func dirPostHandler(w http.ResponseWriter, r *http.Request, repo Repo) {
	s := r.FormValue("submit")
	if s == "Update" {
		readmePath := filepath.Join(repo.rp, "README.md")
		f, err := os.Create(readmePath)
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
		createPath := filepath.Join(repo.rp, name)
		if ForD == "File" {
			if filepath.Ext(createPath) == "" {
				createPath += ".md"
			}
			_, err = os.OpenFile(createPath, os.O_CREATE, 0644)
			if err != nil {
				log.Println(err, "Cannot create file")
			}
		} else if ForD == "Dir" {
			err = os.Mkdir(createPath, 0755)
			if err != nil {
				log.Println(err, "Cannot create directory")
			}
		}

		updateDirTree()
	}
	http.Redirect(w, r, repo.vp, http.StatusFound)
}

func filePostHandler(w http.ResponseWriter, r *http.Request, repo Repo) {
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
}

func repoPostHandler(w http.ResponseWriter, r *http.Request) {
	repo := initRepo(r)

	// Save Display
	if repo.act == "S" {
		// check whether file or directory
		f, err := os.Stat(repo.rp)
		if err != nil {
			log.Println(err, "Failure to checking if file exists")
		}
		if f.IsDir() {
			dirPostHandler(w, r, repo)
		} else {
			filePostHandler(w, r, repo)
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
				"url_for":  func(path string) string { return protocol + baseurl + path },
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
	p.HandleFunc("/{path:.*}", repoHandler).Methods("GET")
	p.HandleFunc("/{path:.*}", repoPostHandler).Methods("POST")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	log.Fatal(http.ListenAndServe(":8080", r))
}
