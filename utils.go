package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"strconv"

	"github.com/BurntSushi/toml"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	_, err := toml.DecodeFile(confFile, &config)
	if err != nil {
		log.Println(err, "Cannot decode toml file")
	}
	// TODO: config のエラーチェック

	createDirTree(&dirTree, config.RepoName)
}

// create Directory Tree
func createDirTree(content *string, current string) {
	ulClass := "class=\"nav nav-pills flex-column\""
	liClass := "class=\"nav-item\""
	divClass := "class=\"nav-link active\""
	aClass := "class=\"nav-link\""
	spanClass := "class=\"sr-only\""
	dir, err := ioutil.ReadDir(current)
	if err != nil {
		log.Println(err, "Cannot read file")
	}

	// start: <ul>
	*content += fmt.Sprintf("<ul %s>\n", ulClass)

	// only "/"
	if current == config.RepoName {
		*content += fmt.Sprintf("<li %s>\n<div %s>Directory Tree<span %s>(current)</span></div>\n</li>", liClass, divClass, spanClass)
	}

	rp := strings.Replace(current, config.RepoName, config.SubDir, -1)
	for _, f := range dir {
		// Don't show ".git", "README.md"
		if f.Name() == ".git" || f.Name() == "README.md" {
			continue
		}

		// Delete file extension
		showName := f.Name()
		if filepath.Ext(showName) != "" {
			showName = strings.TrimSuffix(showName, filepath.Ext(showName))
		}

		*content += fmt.Sprintf("<li %s><a %s href=\"%s\">%s</a></li>\n", liClass, aClass, scheme+path.Join(config.BaseURL, rp, f.Name()), showName)
		fInfo, err := os.Stat(filepath.Join(current, f.Name()))
		if err != nil {
			log.Println(err, "Cannot check file info")
		}
		if fInfo.IsDir() {
			createDirTree(content, path.Join(current, f.Name()))
		}
	}
	*content += fmt.Sprintf("</ul>\n")
	// end: </ul>
}

func updateDirTree() {
	dirTree = ""
	createDirTree(&dirTree, config.RepoName)
}

func createLinkPath(p string) string {
	var linkPath []string
	for p != "/" {
		if strings.TrimPrefix(p, "/") == config.SubDir {
			linkPath = append([]string{fmt.Sprintf("<a href=\"%s\">%s</a>\n", scheme+path.Join(config.BaseURL, p), "Top")}, linkPath...)
		} else {
			linkPath = append([]string{fmt.Sprintf("<a href=\"%s\">%s</a>\n", scheme+path.Join(config.BaseURL, p), path.Base(p))}, linkPath...)
		}
		p = path.Dir(p)
	}
	return strings.Join(linkPath, " / \n")
}

func GetRealRepoPath(rp string) string {
	if rp == config.RepoName {
		return "."
	}

	if strings.HasPrefix(rp, config.RepoName) {
		rpArray := strings.Split(rp, "/")
		return strings.Join(rpArray[1:], "/")
	}
	return rp
}

// git add + git commit
func gitCommit(p string) {
	prev, err := filepath.Abs(".")
	if err != nil {
		log.Println(err)
	}
	defer os.Chdir(prev)

	err = os.Chdir(config.RepoName)
	if err != nil {
		log.Println(err, "Failed to exec \"cd "+config.RepoName+"\"")
	}

	gitPath := GetRealRepoPath(p)
	fmt.Println(gitPath)
	out, err := exec.Command("git", "add", gitPath).Output()
	if err != nil {
		log.Println(err, "Failed to exec \"git add "+gitPath+"\"")
	}
	fmt.Println(string(out), err)

	now := time.Now().Format("2006-01-02_15:04:05")
	out, err = exec.Command("git", "commit", "-m", now).Output()
	if err != nil {
		log.Println(err, "Failed to exec \"git commit -m "+now+"\"")
	}
	fmt.Println(string(out))
}

// git log
func gitLog(p string) []CommitLog {
	var commitList []CommitLog

	prev, err := filepath.Abs(".")
	if err != nil {
		log.Println(err)
	}
	defer os.Chdir(prev)

	err = os.Chdir(config.RepoName)
	if err != nil {
		log.Println(err, "Failed to exec \"cd "+config.RepoName+"\"")
	}

	gitPath := GetRealRepoPath(p)
	linecount, err := strconv.Atoi(config.DiffLines)
	if err != nil {
		log.Println(err, "Cannot convert string to int")
	}
	out, err := exec.Command("git", "log", "-"+fmt.Sprint(linecount), "--pretty=format:\"%s"+config.DiffSeparator+"%h\"", gitPath).Output()
	if err != nil {
		log.Println(err, "Failed to exec \"git diff\"")
	}

	list := strings.Split(string(out), "\n")
	for i, v := range list {
		// ignore moust recent commit
		if i == 0 {
			continue
		}
		v = v[1 : len(v)-1]
		one := strings.Split(v, config.DiffSeparator)
		commitList = append(commitList, CommitLog{Name: one[0], Hash: one[1]})
	}
	return commitList
}

// git diff
func gitDiff(p string, hash string) string {
	prev, err := filepath.Abs(".")
	if err != nil {
		log.Println(err)
	}
	defer os.Chdir(prev)

	err = os.Chdir(config.RepoName)
	if err != nil {
		log.Println(err, "Failed to exec \"cd "+config.RepoName+"\"")
	}

	gitPath := GetRealRepoPath(p)
	out, err := exec.Command("git", "diff", hash, gitPath).Output()
	if err != nil {
		log.Println(err, "Failed to exec \"git diff\"")
	}
	return string(out)
}
