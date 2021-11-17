package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"

	"github.com/charlieegan3/photos/cms/internal/pkg/server/handlers/admin"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/handlers/devices"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
)

func Serve(addr, port, adminUsername, adminPassword string, db *sql.DB, bucket *blob.Bucket, renderer templating.PageRenderer) {
	router := mux.NewRouter()
	router.Use(InitMiddlewareLogging())
	router.Use(InitMiddleware404())

	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(InitMiddlewareAuth(adminUsername, adminPassword))

	adminRouter.HandleFunc("", admin.BuildAdminIndexHandler(renderer)).Methods("GET")
	adminRouter.HandleFunc("/devices", devices.BuildIndexHandler(db, renderer)).Methods("GET")
	adminRouter.HandleFunc("/devices", devices.BuildCreateHandler(db, bucket, renderer)).Methods("POST")
	adminRouter.HandleFunc("/devices/new", devices.BuildNewHandler(renderer)).Methods("GET")
	adminRouter.HandleFunc("/devices/{deviceSlug}", devices.BuildGetHandler(db, renderer)).Methods("GET")
	// handles update and delete
	adminRouter.HandleFunc("/devices/{deviceSlug}", devices.BuildFormHandler(db, bucket, renderer)).Methods("POST")

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
