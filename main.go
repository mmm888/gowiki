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
	"strings"

	"github.com/gorilla/mux"
	"github.com/russross/blackfriday"
	"github.com/unrolled/render"
)

var (
	re       *render.Render
	reponame string
	cd       string
	path     []string
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
	reponame = rname
	http.Redirect(w, r, "/repo", http.StatusFound)
}

func dirHandler(w http.ResponseWriter, repo Repo) {
	switch repo.act {
	/* Edit Display */
	case "E":
		fmt.Println("Edit")
	/* Save Display */
	case "S":
		fmt.Println("Save")
	/* Show Display */
	default:
		var files []string
		dir, err := ioutil.ReadDir(repo.rp)
		if err != nil {
			log.Println(err, "Cannot read file")
		}
		for _, f := range dir {
			files = append(files, f.Name())
		}
		err = re.HTML(w, http.StatusOK, "repo_dir", struct {
			Files []string
			Path  string
			Epath string
			Spath string
		}{
			files, repo.vp + "/", repo.evp, repo.svp,
		})
		if err != nil {
			log.Println(err, "Cannot generate template")
		}
	}
}

func fileHandler(w http.ResponseWriter, r *http.Request, repo Repo) {
	switch repo.act {
	/* Edit Display */
	case "E":
		file, err := ioutil.ReadFile(repo.rp)
		if err != nil {
			log.Println(err, "Cannot read file")
		}
		err = re.HTML(w, http.StatusOK, "edit_file", struct {
			Content string
			Path    string
			Epath   string
			Spath   string
		}{
			string(file), repo.vp, repo.evp, repo.svp,
		})
		if err != nil {
			log.Println(err, "Cannot generate template")
		}
	/* Save Display */
	case "S":
		s := r.FormValue("submit")
		if s == "Save" {
			con := r.FormValue("content")
			f, err := os.Create(repo.rp)
			if err != nil {
				log.Println(err, "Cannot create file")
			}
			defer f.Close()

			_, err = f.Write([]byte(con))
			if err != nil {
				log.Println(err, "Cannot writer file")
			}
		}
		http.Redirect(w, r, repo.vp, http.StatusFound)
	/* Show Display */
	default:
		file, err := ioutil.ReadFile(repo.rp)
		if err != nil {
			log.Println(err, "Cannot read file")
		}
		file_md := blackfriday.MarkdownCommon(file)
		err = re.HTML(w, http.StatusOK, "repo_file", struct {
			Content string
			Path    string
			Epath   string
			Spath   string
		}{
			string(file_md), repo.vp, repo.evp, repo.svp,
		})
		if err != nil {
			log.Println(err, "Cannot generate template")
		}
	}
}

func repoHandler(w http.ResponseWriter, r *http.Request) {
	var repo Repo
	actEdit := "?action=E"
	actSave := "?action=S"
	repo.act = r.FormValue("action")
	path := mux.Vars(r)["path"]
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
	f, err := os.Stat(repo.rp)
	if err != nil {
		log.Println(err, "Failure to checking if file exists")
	}

	if f.IsDir() {
		dirHandler(w, repo)
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

func ErrorPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", "<p>Internal Server Error</p>")
}

/*
type directory struct {
	files []string
}

var dtree []directory
*/

func TestHandler(w http.ResponseWriter, r *http.Request) {
	/*
		tmpl := template.Must(template.ParseFiles("./test.html"))
		tmpl.Execute(w, nil)
	*/
	var files []string
	dir, err := ioutil.ReadDir(reponame)
	if err != nil {
		log.Println(err, "Cannot read file")
	}
	for _, f := range dir {
		files = append(files, f.Name())
	}
	err = re.HTML(w, http.StatusOK, "dirtree", struct {
		Files []string
		Path  string
	}{
		files, r.URL.String(),
	})
	if err != nil {
		log.Println(err, "Cannot generate template")
	}
}

func main() {
	baseurl := "http://dev01-xenial:8080"
	reponame = "wikitest/"

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
	p := r.PathPrefix("/repo/").Subrouter()
	p.HandleFunc("/{path:.*}", repoHandler)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	log.Fatal(http.ListenAndServe(":8080", r))
}
