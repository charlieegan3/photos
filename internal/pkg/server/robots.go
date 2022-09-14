package server

import (
	_ "embed"
	"fmt"
	"net/http"
)

//go:embed static/robots.txt
var robotsContent string

func buildRobotsHandler() (handler func(http.ResponseWriter, *http.Request)) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=60")
		w.Header().Set("Content-Type", "text/plain")

		fmt.Fprint(w, robotsContent)
	}
}
