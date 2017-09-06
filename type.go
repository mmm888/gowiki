package main

// Config is decoded toml file
type Config struct {
	Protocol string
	BaseURL  string
	RepoName string
	SubDir   string
}

// Repo is common part for "/repo"
type Repo struct {
	act string
	rp  string
	vp  string
	evp string
	svp string
}
