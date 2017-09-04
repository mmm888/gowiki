package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	dirTree(&dirtree, reponame)
}

/* create Directory Tree */
func dirTree(content *string, current string) {
	ulClass := "class=\"nav nav-pills flex-column\""
	liClass := "class=\"nav-item\""
	divClass := "class=\"nav-link active\""
	aClass := "class=\"nav-link\""
	spanClass := "class=\"sr-only\""
	dir, err := ioutil.ReadDir(current)
	if err != nil {
		log.Println(err, "Cannot read file")
	}
	/* <ul> */
	*content += fmt.Sprintf("<ul %s>\n", ulClass)
	if current == reponame {
		/* top tree */
		*content += fmt.Sprintf("<li %s>\n<div %s>Directory Tree<span %s>(current)</span></div>\n</li>", liClass, divClass, spanClass)
	}
	rp := strings.Replace(current, reponame, subdir, -1)
	for _, f := range dir {
		if f.Name() == ".git" || f.Name() == "README.md" {
			continue
		}
		*content += fmt.Sprintf("<li %s><a %s href=\"%s\">%s</a></li>\n", liClass, aClass, baseurl+"/"+rp+f.Name(), f.Name())
		fInfo, err := os.Stat(current + f.Name())
		if err != nil {
			log.Println(err, "Cannot check file info")
		}
		if fInfo.IsDir() {
			dirTree(content, current+f.Name()+"/")
		}
	}
	*content += fmt.Sprintf("</ul>\n")
	/* </ul> */
}
