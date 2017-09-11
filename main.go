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
	dirTree  string
	config   Config
	confFile = "config.toml"
	scheme   = "http://"
	actEdit  = "?action=E"
	actSave  = "?action=S"
)

// RootHandler is routing of "/"
func RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello")
}

//tmp := gitLog(repo.rp)
//fmt.Println(gitDiff(repo.rp, tmp[0]))

func diffListHandler(w http.ResponseWriter, r *http.Request) {
	p := r.FormValue("path")
	if p == "" {
		p = config.RepoName
	}
	commitList := gitLog(p)
	err := re.HTML(w, http.StatusOK, "diff_list", struct {
		CommitList     []CommitLog
		IsHeaderOption bool
		Path           string
		File           string
	}{
		commitList,
		false,
		GetRealRepoPath(config.RepoName),
		p,
	})
	if err != nil {
		log.Println(err, "Cannot generate template")
	}
}

func diffShowHandler(w http.ResponseWriter, r *http.Request) {
	h := mux.Vars(r)["hash"]
	p := r.FormValue("path")
	if p == "" {
		p = config.RepoName
	}
	con := gitDiff(p, h)

	defaultRows := 20
	// TODO: 行数によってrowsの大きさを変える
	//	fmt.Println(len(strings.Split(con, "\n")))
	err := re.HTML(w, http.StatusOK, "diff_show", struct {
		Content        string
		Path           string
		Rows           string
		IsHeaderOption bool
	}{
		con,
		GetRealRepoPath(config.RepoName),
		fmt.Sprint(defaultRows),
		false,
	})
	if err != nil {
		log.Println(err, "Cannot generate template")
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
	config.RepoName = rname
	http.Redirect(w, r, "/repo", http.StatusFound)
}

func initRepo(r *http.Request) Repo {
	var repo Repo
	p := mux.Vars(r)["path"]
	repo.act = r.FormValue("action")
	repo.rp = filepath.Join(config.RepoName, p)
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
	var con, tmplname string
	switch repo.act {

	// Edit Display
	case "E":
		readmePath := filepath.Join(repo.rp, "README.md")
		f, err := ioutil.ReadFile(readmePath)
		if err != nil {
			log.Println(err, "Cannot read file")
		}
		con = string(f)

		tmplname = "edit_dir"

	// Show Display
	default:
		_, err := os.Stat(repo.rp + "/README.md")
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
		con = string(md)

		tmplname = "repo"
	}

	// only Edit: Spath
	// only Show: Dirtree, LinkPath
	err := re.HTML(w, http.StatusOK, tmplname, struct {
		Content        string
		Path           string
		Epath          string
		Spath          string
		Dirtree        string
		LinkPath       string
		IsHeaderOption bool
	}{
		con, GetRealRepoPath(repo.rp), repo.evp, repo.svp, dirTree, createLinkPath(repo.vp), true,
	})
	if err != nil {
		log.Println(err, "Cannot generate template")
	}
}

func fileHandler(w http.ResponseWriter, r *http.Request, repo Repo) {
	var con, tmplname string
	switch repo.act {

	// Edit Display
	case "E":
		f, err := ioutil.ReadFile(repo.rp)
		if err != nil {
			log.Println(err, "Cannot read file")
		}
		con = string(f)

		tmplname = "edit_file"

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
		con = string(md)
		tmplname = "repo"
	}

	// only Edit: Spath, FileName
	// only Show: Dirtree, LinkPath
	err := re.HTML(w, http.StatusOK, tmplname, struct {
		Content        string
		Path           string
		Epath          string
		Spath          string
		FileName       string
		Dirtree        string
		LinkPath       string
		IsHeaderOption bool
	}{
		con, GetRealRepoPath(repo.rp), repo.evp, repo.svp, filepath.Base(repo.rp), dirTree, createLinkPath(repo.vp), true,
	})
	if err != nil {
		log.Println(err, "Cannot generate template")
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
			log.Println(err, "Cannot write file")
		}

		name := r.FormValue("FileName")
		ForD := r.FormValue("ForD")
		createPath := filepath.Join(repo.rp, filepath.Base(name))
		if ForD == "File" {
			if filepath.Ext(createPath) == "" {
				name += ".md"
				createPath += ".md"
			}
			_, err = os.OpenFile(createPath, os.O_CREATE, 0644)
			if err != nil {
				log.Println(err, "Cannot create file")
			}

			gitCommit(createPath)
		} else if ForD == "Dir" {
			err = os.Mkdir(createPath, 0755)
			if err != nil {
				log.Println(err, "Cannot create directory")
			}
		}

		if ForD != "None" {
			updateDirTree()
			http.Redirect(w, r, filepath.Join(repo.vp, filepath.Base(name)), http.StatusFound)
		}
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

		gitCommit(repo.rp)
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
				"url_for":  func(path string) string { return scheme + config.BaseURL + path },
				"safehtml": func(text string) template.HTML { return template.HTML(text) },
				"stradd":   func(a string, b string) string { return a + b },
				"difflink": func(p string) string {
					if p == "" || p == "." {
						return ""
					}
					return "?path=" + p
				},
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
	r.HandleFunc("/diff", diffListHandler)
	r.HandleFunc("/diff/{hash}", diffShowHandler)

	r.HandleFunc("/repo", repoHandler).Methods("GET")
	r.HandleFunc("/repo", repoPostHandler).Methods("POST")
	p := r.PathPrefix("/repo").Subrouter()
	p.HandleFunc("/{path:.*}", repoHandler).Methods("GET")
	p.HandleFunc("/{path:.*}", repoPostHandler).Methods("POST")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	log.Fatal(http.ListenAndServe(":8080", r))
}
