package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"

	"github.com/charlieegan3/photos/cms/internal/pkg/server/handlers"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/handlers/admin"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/handlers/admin/devices"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/handlers/admin/locations"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/handlers/admin/medias"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/handlers/admin/posts"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/handlers/admin/tags"
	publicdevices "github.com/charlieegan3/photos/cms/internal/pkg/server/handlers/public/devices"
	publiclocations "github.com/charlieegan3/photos/cms/internal/pkg/server/handlers/public/locations"
	publicmedias "github.com/charlieegan3/photos/cms/internal/pkg/server/handlers/public/medias"
	publicposts "github.com/charlieegan3/photos/cms/internal/pkg/server/handlers/public/posts"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
)

func Serve(
	environment, hostname, addr, port, adminUsername, adminPassword string,
	db *sql.DB,
	bucket *blob.Bucket,
	mapServerURL, mapServerAPIKey string,
	renderer templating.PageRenderer,
	rendererAdmin templating.PageRenderer,
) {
	router := mux.NewRouter()
	router.Use(InitMiddlewareLogging())
	// router.Use(InitMiddlewareHTTPS(hostname, environment))
	router.NotFoundHandler = http.HandlerFunc(notFound)

	stylesHandler, err := buildStylesHandler()
	if err != nil {
		log.Fatalf("failed to build styles handler: %s", err)
	}
	router.HandleFunc("/styles.css", stylesHandler).Methods("GET")

	router.HandleFunc("/favicon.ico", faviconHandler).Methods("GET")

	router.HandleFunc("", handlers.BuildRedirectHandler("/")).Methods("GET")
	router.HandleFunc("/", publicposts.BuildIndexHandler(db, renderer)).Methods("GET")
	router.HandleFunc("/posts/period/{from}-to-{to}", publicposts.BuildPeriodHandler(db, renderer)).Methods("GET")
	router.HandleFunc("/posts/period/{from}", publicposts.BuildPeriodHandler(db, renderer)).Methods("GET")
	router.HandleFunc("/posts/{postID}", publicposts.BuildGetHandler(db, renderer)).Methods("GET")
	router.HandleFunc("/posts/", handlers.BuildRedirectHandler("/")).Methods("GET")
	router.HandleFunc("/posts", handlers.BuildRedirectHandler("/")).Methods("GET")
	router.HandleFunc("/locations/{locationID}/map.jpg", publiclocations.BuildMapHandler(db, bucket, mapServerURL, mapServerAPIKey)).Methods("GET")
	router.HandleFunc("/medias/{mediaID}/{file}.{kind}", publicmedias.BuildMediaHandler(db, bucket)).Methods("GET")
	router.HandleFunc("/devices/{deviceID}/icon.{kind}", publicdevices.BuildIconHandler(db, bucket)).Methods("GET")

	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(InitMiddlewareAuth(adminUsername, adminPassword))

	adminRouter.HandleFunc("", admin.BuildAdminIndexHandler(rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/", handlers.BuildRedirectHandler("/admin")).Methods("GET")
	adminRouter.HandleFunc("/devices", devices.BuildIndexHandler(db, rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/devices", devices.BuildCreateHandler(db, bucket, rendererAdmin)).Methods("POST")
	adminRouter.HandleFunc("/devices/new", devices.BuildNewHandler(rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/devices/{deviceID}", devices.BuildGetHandler(db, rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/devices/{deviceID}", devices.BuildFormHandler(db, bucket, rendererAdmin)).Methods("POST")

	adminRouter.HandleFunc("/tags", tags.BuildIndexHandler(db, rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/tags", tags.BuildCreateHandler(db, rendererAdmin)).Methods("POST")
	adminRouter.HandleFunc("/tags/new", tags.BuildNewHandler(rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/tags/{tagName}", tags.BuildGetHandler(db, rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/tags/{tagName}", tags.BuildFormHandler(db, rendererAdmin)).Methods("POST")

	adminRouter.HandleFunc("/locations", locations.BuildIndexHandler(db, rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/locations", locations.BuildCreateHandler(db, rendererAdmin)).Methods("POST")
	adminRouter.HandleFunc("/locations/new", locations.BuildNewHandler(rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/locations/{locationID}", locations.BuildGetHandler(db, rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/locations/{locationID}", locations.BuildFormHandler(db, rendererAdmin)).Methods("POST")

	adminRouter.HandleFunc("/medias", medias.BuildIndexHandler(db, rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/medias", medias.BuildCreateHandler(db, bucket, rendererAdmin)).Methods("POST")
	adminRouter.HandleFunc("/medias/new", medias.BuildNewHandler(db, rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/medias/{mediaID}", medias.BuildGetHandler(db, rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/medias/{mediaID}", medias.BuildFormHandler(db, bucket, rendererAdmin)).Methods("POST")

	adminRouter.HandleFunc("/posts", posts.BuildIndexHandler(db, rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/posts", posts.BuildCreateHandler(db, rendererAdmin)).Methods("POST")
	adminRouter.HandleFunc("/posts/new", posts.BuildNewHandler(db, rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/posts/{postID}", posts.BuildGetHandler(db, rendererAdmin)).Methods("GET")
	adminRouter.HandleFunc("/posts/{postID}", posts.BuildFormHandler(db, rendererAdmin)).Methods("POST")

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("%s:%s", addr, port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
