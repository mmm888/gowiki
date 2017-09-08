package main

// Config is decoded toml file
type Config struct {
	Scheme   string
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

type gLog struct {
	name string
	hash string
}
