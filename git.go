package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

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
	if gitPath == "" {
		gitPath = "."
	}
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
	if gitPath == "" {
		gitPath = "."
	}
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
	if gitPath == "" {
		gitPath = "."
	}
	out, err := exec.Command("git", "diff", hash, gitPath).Output()
	if err != nil {
		log.Println(err, "Failed to exec \"git diff\"")
	}
	return string(out)
}
