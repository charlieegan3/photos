package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	devices "github.com/charlieegan3/photos/cms/internal/pkg/server/handlers"
)

func Serve(addr, port, adminUsername, adminPassword string, db *sql.DB) {
	router := mux.NewRouter()
	router.Use(InitMiddlewareLogging())
	router.Use(InitMiddleware404())

	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(InitMiddlewareAuth(adminUsername, adminPassword))

	adminRouter.HandleFunc("/devices", devices.BuildIndexHandler(db)).Methods("GET")
	adminRouter.HandleFunc("/devices", devices.BuildCreateHandler(db)).Methods("POST")
	adminRouter.HandleFunc("/devices/new", devices.BuildNewHandler()).Methods("GET")
	adminRouter.HandleFunc("/devices/{deviceName}", devices.BuildGetHandler(db)).Methods("GET")
	// handles update and delete
	adminRouter.HandleFunc("/devices/{deviceName}", devices.BuildFormHandler(db)).Methods("POST")

	router.NotFoundHandler = http.HandlerFunc(notFound)

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("%s:%s", addr, port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func notFound(w http.ResponseWriter, r *http.Request) {
	logger := logrus.New()
	entry := logrus.NewEntry(logger)
	entry.Info("not found", r.URL.Path)
}
