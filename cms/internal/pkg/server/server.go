package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	devices "github.com/charlieegan3/photos/cms/internal/pkg/server/handlers"
	"github.com/gorilla/mux"
)

func Serve(addr, port, adminUsername, adminPassword string, db *sql.DB) {
	r := mux.NewRouter()

	adminRouter := r.PathPrefix("/admin").Subrouter()
	adminRouter.Use(InitMiddlewareAuth(adminUsername, adminPassword))

	adminRouter.HandleFunc("/devices", devices.BuildIndexHandler(db))

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", addr, port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
