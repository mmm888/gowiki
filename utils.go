package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

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
			showName = strings.TrimRight(showName, filepath.Ext(showName))
		}

		*content += fmt.Sprintf("<li %s><a %s href=\"%s\">%s</a></li>\n", liClass, aClass, config.Protocol+path.Join(config.BaseURL, rp, f.Name()), showName)
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
		if strings.TrimLeft(p, "/") == config.SubDir {
			linkPath = append([]string{fmt.Sprintf("<a href=\"%s\">%s</a>\n", config.Protocol+path.Join(config.BaseURL, p), "Top")}, linkPath...)
		} else {
			linkPath = append([]string{fmt.Sprintf("<a href=\"%s\">%s</a>\n", config.Protocol+path.Join(config.BaseURL, p), path.Base(p))}, linkPath...)
		}
		p = path.Dir(p)
	}
	return strings.Join(linkPath, " / \n")
}
