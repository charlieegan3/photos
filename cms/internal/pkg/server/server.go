package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func Serve(addr, port, adminUsername, adminPassword string) {
	r := mux.NewRouter()

	adminRouter := r.PathPrefix("/admin").Subrouter()
	adminRouter.Use(InitMiddlewareAuth(adminUsername, adminPassword))

	r.HandleFunc("/public", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "public")
	})
	adminRouter.HandleFunc("/secret", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "secret")
	})

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", addr, port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
