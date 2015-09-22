package web

import (
	"log"
	"net/http"
)

type WebFace struct {
	Addr            string
	StaticRoot      string
	Templates       string
	Router          *http.ServeMux
	RedirectHandler http.HandlerFunc

	OutMsg chan string
	InMsg  chan string
}

func MakeWebFace(addr string, static_root string, templatesFolder string) *WebFace {
	wf := &WebFace{
		Addr:       addr,
		Router:     http.NewServeMux(),
		StaticRoot: static_root,
		Templates:  templatesFolder,

		OutMsg: make(chan string),
		InMsg:  make(chan string),
	}

	wf.Router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(static_root))))

	go wf.HostLoop()

	return wf
}

func (wf *WebFace) HostLoop() {
	defer log.Println("Stopped Listening")

	log.Println("Listening on", wf.Addr)
	err := http.ListenAndServe(wf.Addr, wf)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func (wf *WebFace) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if wf.RedirectHandler != nil {
		wf.RedirectHandler(rw, req)
		return
	}

	log.Println(req.URL)

	wf.Router.ServeHTTP(rw, req)
}

//===
