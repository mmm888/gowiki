package main

import "log"

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

/* logging */
func logging(args ...interface{}) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println(args...)
}
