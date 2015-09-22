package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"

	"./web"
)

var (
	reDayPathMatch  = regexp.MustCompile("/day/([0-9]+)[/\\-]([0-9]+)[/\\-]([0-9]+)")
	reFilePathMatch = regexp.MustCompile("/file/([^/]+)")
)

const (
	dateFormat     = "2006-01-02"
	dateFormatLong = "2006-01-02T15:04:05.000Z"
)

func SetupWebFace(wf *web.WebFace) {
	sh := SummaryHandle{}

	wf.Router.Handle("/", sh)
}

////////////////////////////////////////////////////////////////////////////////
// Summary Handle
type SummaryHandle struct {
}

func (sh SummaryHandle) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	fmt.Fprintln(rw, "Nothing to see here")
	sumTemp, err := template.ParseFiles("summary.html")
	if err != nil {
		log.Fatalln("Error parsing:", err)
	}
	e := sumTemp.Execute(rw, sh)
	if e != nil {
		log.Println("Error in Temp", e)
	}
}
