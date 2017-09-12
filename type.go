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
}

func (repo Repo) GetActPath(a string) string {
	return GetActPath(repo.vp, a)
}

func (repo Repo) GetRealRepoPath() string {
	return GetRealRepoPath(repo.rp)
}

type CommitLog struct {
	Name string
	Hash string
}
