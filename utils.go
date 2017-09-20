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
	// TODO: config のエラーチェック

	createDirTree(&dirTree, config.RepoName)
}

// create Directory Tree
func createDirTree(content *string, current string) {
	ulClass := "style=\"display: none;\""
	liClass := "class=\"folder\""
	//	divClass := "class=\"active\""
	//	spanClass := "class=\"sr-only\""

	dir, err := ioutil.ReadDir(current)
	if err != nil {
		log.Println(err, "Cannot read file")
	}

	// start: <div>
	// only "/"
	if current == config.RepoName {
		*content += fmt.Sprintf("<div id=\"tree\" class=\"tree-body\">\n")
		*content += fmt.Sprintf("<ul %s>\n", ulClass)
		//		*content += fmt.Sprintf("<li>\n<div %s>Directory Tree<span %s>(current)</span></div>\n</li>", divClass, spanClass)
	} else {
		*content += fmt.Sprintf("<ul>\n")
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

		fInfo, err := os.Stat(filepath.Join(current, f.Name()))
		if err != nil {
			log.Println(err, "Cannot check file info")
		}

		// which Directory or File
		if fInfo.IsDir() {
			*content += fmt.Sprintf("<li %s><a target=\"content\" href=\"%s\">%s</a><\n", liClass, GetFullPath(rp, f.Name()), showName)
			createDirTree(content, path.Join(current, f.Name()))
		} else {
			*content += fmt.Sprintf("<li><a target=\"content\" href=\"%s\">%s</a>\n", GetFullPath(rp, f.Name()), showName)
		}
	}
	*content += fmt.Sprintf("</ul>\n")
	if current == config.RepoName {
		*content += fmt.Sprintf("</div>\n")
	}
	// end: </div>
}

func updateDirTree() {
	dirTree = ""
	createDirTree(&dirTree, config.RepoName)
}

func createLinkPath(p string) string {
	var linkPath []string
	for p != "." {
		if p == config.SubDir {
			linkPath = append([]string{fmt.Sprintf("<a href=\"%s\">%s</a>\n", GetFullPath(p), "Top")}, linkPath...)
		} else {
			linkPath = append([]string{fmt.Sprintf("<a href=\"%s\">%s</a>\n", GetFullPath(p), path.Base(p))}, linkPath...)
		}
		p = path.Dir(p)
	}
	return strings.Join(linkPath, " / \n")
}

func GetRealRepoPath(rp string) string {
	if rp == config.RepoName {
		return ""
	}

	if strings.HasPrefix(rp, config.RepoName) {
		rpArray := strings.Split(rp, "/")
		return strings.Join(rpArray[1:], "/")
	}
	return rp
}

func GetNoActPath(p string) string {
	action := "?action="

	index := strings.LastIndex(p, action)
	// repo.vp の Suffix が "?action=." かの判定
	if index != -1 && p[index:len(p)-1] == action {
		p = p[:index]
	}

	return p
}

func GetActPath(p, key string) string {
	action := "?action="

	index := strings.LastIndex(p, action)
	// repo.vp の Suffix が "?action=." かの判定
	if index != -1 && p[index:len(p)-1] == action {
		p = p[:index]
	}

	return p + action + key
}

func GetFullPath(p ...string) string {
	urlpath := config.BaseURL
	for _, v := range p {
		urlpath = path.Join(urlpath, v)
	}
	return scheme + urlpath

}
