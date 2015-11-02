/*  JOB: Upload Images to Drive

*/
package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	google "./google"
	web "./web"
)

func startUploader(wf *web.WebFace) {
	wf.RedirectHandler = func(rw http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(rw, "Starting Server on %s", *addr)
	}

	// Login
	log.Println("Login")
	Tok, cErr := google.Login(wf, google.GetClientScope())
	if cErr != nil {
		log.Fatalln("Login Error:", cErr)
	}

	iTok, iErr := google.GetIdentity(Tok)
	if iErr != nil {
		log.Fatalln("Identity Error:", iErr)
	}
	fmt.Println(iTok)

	b := new(bytes.Buffer)
	google.EncodeToken(Tok, b)

	wf.RedirectHandler = nil

	// Now Scan and upload

	// google.InsertFile("Camera B", "", "", "image/webp", respB.Body)
}
