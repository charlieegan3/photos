package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	devices "github.com/charlieegan3/photos/cms/internal/pkg/server/handlers"
)

func Serve(addr, port, adminUsername, adminPassword string, db *sql.DB) {
	router := mux.NewRouter()
	router.Use(InitMiddlewareLogging())

	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(InitMiddlewareAuth(adminUsername, adminPassword))

	adminRouter.HandleFunc("/devices", devices.BuildIndexHandler(db)).Methods("GET")
	adminRouter.HandleFunc("/devices", devices.BuildCreateHandler(db)).Methods("POST")
	adminRouter.HandleFunc("/devices/new", devices.BuildNewHandler()).Methods("GET")

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("%s:%s", addr, port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
