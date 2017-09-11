package main

// Config is decoded toml file
type Config struct {
	BaseURL       string
	RepoName      string
	SubDir        string
	DiffLines     string
	DiffSeparator string
}

// Repo is common part for "/repo"
type Repo struct {
	act string
	rp  string
	vp  string
	evp string
	svp string
}

type CommitLog struct {
	Name string
	Hash string
}
