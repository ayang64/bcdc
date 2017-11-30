package main

import (
	"flag"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type App struct {
	debug       *log.Logger
	fileServer  http.Handler
	imageServer http.Handler
	tmpl        *template.Template
}

func (a App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.debug.Printf("handling path %q", r.URL.Path)

	src := r.URL.Path

	if len(src) > 1 && src[len(src)-1] == '/' {
		// strip trailing slash and redirect
		dst := src[:len(src)-1]
		a.debug.Printf("dst: %q", dst)
		http.Redirect(w, r, dst, 307)
		return
	}

	switch {
	case strings.HasPrefix(r.URL.Path, "/images/"):
		a.debug.Printf("handling static file in assets directory: %q", r.URL.Path)
		a.imageServer.ServeHTTP(w, r)

	case strings.HasPrefix(r.URL.Path, "/assets/"):
		a.debug.Printf("handling static file in assets directory: %q", r.URL.Path)
		a.fileServer.ServeHTTP(w, r)

	case r.URL.Path == "/":
		err := a.tmpl.ExecuteTemplate(w, "index.html", nil)

		if err != nil {
			a.debug.Printf("%v", err)
		}

	default:
		http.Redirect(w, r, "/", 307)
	}
}

func main() {
	debug := flag.Bool("debug", false, "Enable debug mode.")
	assets := flag.String("assets", "./assets", "Directory for static assets like images, css, and one day templates.")
	images := flag.String("images", "./images", "Direcotry containging images served from /images/ url.")
	flag.Parse()

	debugWriter := func() io.Writer {
		if *debug == true {
			return os.Stderr
		}
		return ioutil.Discard
	}

	tmpl, err := template.ParseGlob(path.Join(*assets, "templates/*"))

	if err != nil {
		log.Fatalf("%v", err)
	}

	app := &App{
		debug:       log.New(debugWriter(), "DEBUG ", log.LstdFlags|log.Lshortfile),
		fileServer:  http.StripPrefix("/assets", http.FileServer(http.Dir(*assets))),
		imageServer: http.StripPrefix("/images", http.FileServer(http.Dir(path.Join(*images)))),
		tmpl:        tmpl,
	}

	server := http.Server{
		Addr:           ":8080",
		Handler:        app,
		ReadTimeout:    500 * time.Millisecond, // timeout after 1/2 second.
		WriteTimeout:   500 * time.Millisecond, // timeout after 1/2 second.
		MaxHeaderBytes: 1 << 12,                // about 4 megabytes
	}

	log.Fatalf("%s", server.ListenAndServe())
}
