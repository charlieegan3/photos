package server

import (
	"embed"
	"net/http"
	"net/url"
)

//go:embed static/*
var staticContent embed.FS

func buildStaticHandler() (handler func(http.ResponseWriter, *http.Request)) {
	return func(w http.ResponseWriter, req *http.Request) {
		rootedReq := http.Request{
			URL: &url.URL{
				Path: "./static" + req.URL.Path,
			},
		}
		http.FileServer(http.FS(staticContent)).ServeHTTP(w, &rootedReq)
	}
}
