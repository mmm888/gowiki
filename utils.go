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

func dirTree(content *string, current string) {
	dir, err := ioutil.ReadDir(current)
	if err != nil {
		log.Println(err, "Cannot read file")
	}
	*content += fmt.Sprintf("<ul>\n")
	rp := strings.Replace(current, reponame, subdir, -1)
	for _, f := range dir {
		if f.Name() == ".git" {
			continue
		}
		*content += fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", baseurl+"/"+rp+f.Name(), f.Name())
		fInfo, err := os.Stat(current + f.Name())
		if err != nil {
			log.Println(err, "Cannot check file info")
		}
		if fInfo.IsDir() {
			dirTree(content, current+f.Name()+"/")
		}
	}
	*content += fmt.Sprintf("</ul>\n")
}

/* logging */
func logging(args ...interface{}) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println(args...)
}
