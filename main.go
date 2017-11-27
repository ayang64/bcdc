package main

import (
	"flag"
	"fmt"
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
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
	<head>
	</head>
		<body style='background: #303030; margin-left: 10%; margin-right: 10%; color: white;'>
		<div style='position: relative;'>
				<img src='/assets/static/bcdc-logo.png' style='width: 25%;' />
				<br />
				<p>
					Thank you for your support.  We are currently closed for business.
					We will probably re-open to sell our handguard grip tape accessories
					by January.  We <i>might</i> resume selling guns and accessories by
					2019.
				</p>
			</div>
		</div>
		<p>
		</p>
	</body>
</html>`)
	default:
		http.Redirect(w, r, "/", 307)
	}
}

func main() {
	debug := flag.Bool("debug", false, "Enable debug mode.")
	assets := flag.String("assets", "./assets", "Directory for static assets like images, css, and one day templates.")
	flag.Parse()

	debugWriter := func() io.Writer {
		if *debug == true {
			return os.Stderr
		}
		return ioutil.Discard
	}

	app := &App{
		debug:       log.New(debugWriter(), "DEBUG ", log.LstdFlags|log.Lshortfile),
		fileServer:  http.StripPrefix("/assets", http.FileServer(http.Dir(*assets))),
		imageServer: http.StripPrefix("/images", http.FileServer(http.Dir(path.Join(*assets, "images")))),
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
