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

	"github.com/BurntSushi/toml"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	_, err := toml.DecodeFile(confFile, &config)
	if err != nil {
		log.Println(err, "Cannot decode toml file")
	}

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

		*content += fmt.Sprintf("<li %s><a %s href=\"%s\">%s</a></li>\n", liClass, aClass, config.Scheme+path.Join(config.BaseURL, rp, f.Name()), showName)
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
			linkPath = append([]string{fmt.Sprintf("<a href=\"%s\">%s</a>\n", config.Scheme+path.Join(config.BaseURL, p), "Top")}, linkPath...)
		} else {
			linkPath = append([]string{fmt.Sprintf("<a href=\"%s\">%s</a>\n", config.Scheme+path.Join(config.BaseURL, p), path.Base(p))}, linkPath...)
		}
		p = path.Dir(p)
	}
	return strings.Join(linkPath, " / \n")
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

	gitPath := strings.TrimPrefix(p, config.RepoName+"/")
	_, err = exec.Command("git", "add", gitPath).Output()
	if err != nil {
		log.Println(err, "Failed to exec \"git add "+gitPath+"\"")
	}

	now := time.Now().Format("2006-01-02_15:04:05")
	_, err = exec.Command("git", "commit", "-m", now).Output()
	if err != nil {
		log.Println(err, "Failed to exec \"git commit -m "+now+"\"")
	}
}

// git log
func gitLog(p string) []gLog {
	var logList []gLog
	prev, err := filepath.Abs(".")
	if err != nil {
		log.Println(err)
	}
	defer os.Chdir(prev)

	err = os.Chdir(config.RepoName)
	if err != nil {
		log.Println(err, "Failed to exec \"cd "+config.RepoName+"\"")
	}

	gitPath := strings.TrimPrefix(p, config.RepoName+"/")
	out, err := exec.Command("git", "log", "--pretty=format:\"%s %h\"", gitPath).Output()
	if err != nil {
		log.Println(err, "Failed to exec \"git diff\"")
	}

	list := strings.Split(string(out), "\n")
	for _, v := range list {
		v = strings.Trim(v, "\"")
		alog := strings.Split(v, " ")
		logList = append(logList, gLog{name: alog[0], hash: alog[1]})
	}
	return logList
}

// git diff
func gitDiff(p string, l gLog) string {
	prev, err := filepath.Abs(".")
	if err != nil {
		log.Println(err)
	}
	defer os.Chdir(prev)

	err = os.Chdir(config.RepoName)
	if err != nil {
		log.Println(err, "Failed to exec \"cd "+config.RepoName+"\"")
	}

	gitPath := strings.TrimPrefix(p, config.RepoName+"/")
	fmt.Println(l.hash, gitPath)
	out, err := exec.Command("git", "diff", l.hash, gitPath).Output()
	if err != nil {
		log.Println(err, "Failed to exec \"git diff\"")
	}
	return string(out)
}
