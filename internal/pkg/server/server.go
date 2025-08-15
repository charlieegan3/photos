package server

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"

	"github.com/charlieegan3/photos/internal/pkg/server/handlers"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/admin"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/admin/devices"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/admin/lenses"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/admin/locations"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/admin/medias"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/admin/posts"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/admin/tags"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/admin/trips"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/public/menu"
	publictrips "github.com/charlieegan3/photos/internal/pkg/server/handlers/public/trips"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"

	publicdevices "github.com/charlieegan3/photos/internal/pkg/server/handlers/public/devices"
	publicLenses "github.com/charlieegan3/photos/internal/pkg/server/handlers/public/lenses"
	publiclocations "github.com/charlieegan3/photos/internal/pkg/server/handlers/public/locations"
	publicmedias "github.com/charlieegan3/photos/internal/pkg/server/handlers/public/medias"
	publicposts "github.com/charlieegan3/photos/internal/pkg/server/handlers/public/posts"
	publictags "github.com/charlieegan3/photos/internal/pkg/server/handlers/public/tags"
)

// Attach adds all routes to the router, this is used in other projects to run
// an instance of the server.
func Attach(
	router *mux.Router,
	db *sql.DB,
	bucket *blob.Bucket,
	mapServerURL, mapServerAPIKey string,
	adminPath string,
	environment string,
	permittedEmailSuffix string,
) error {
	renderer := templating.BuildPageRenderFunc(true, "")
	rendererMenu := templating.BuildPageRenderFunc(false, "")
	rendererMap := templating.BuildPageRenderFunc(true, publiclocations.HeadContent)
	rendererAdmin := templating.BuildPageRenderFunc(true, "", "admin")

	router.Use(InitMiddlewareLogging())
	router.NotFoundHandler = http.HandlerFunc(notFound)

	stylesHandler, err := buildStylesHandler()
	if err != nil {
		return fmt.Errorf("failed to build styles handler: %w", err)
	}
	router.HandleFunc("/styles.css", stylesHandler).Methods(http.MethodGet)

	router.HandleFunc("/rss.xml", publicposts.BuildRSSHandler(db)).Methods(http.MethodGet)

	router.HandleFunc("", handlers.BuildRedirectHandler("/")).Methods(http.MethodGet)
	router.HandleFunc("/", publicposts.BuildIndexHandler(db, renderer)).Methods(http.MethodGet)

	router.HandleFunc("/menu", menu.BuildIndexHandler(db, rendererMenu)).Methods(http.MethodGet)

	router.HandleFunc("/tags", publictags.BuildIndexHandler(db, renderer)).Methods(http.MethodGet)
	router.HandleFunc("/tags/{tagName}", publictags.BuildGetHandler(db, renderer)).Methods(http.MethodGet)

	router.HandleFunc("/posts/latest.json", publicposts.BuildLatestHandler(db)).Methods(http.MethodGet)
	router.HandleFunc("/posts/period/{from}-to-{to}", publicposts.BuildPeriodHandler(db, renderer)).Methods(http.MethodGet)
	router.HandleFunc("/posts/period/{from}", publicposts.BuildPeriodHandler(db, renderer)).Methods(http.MethodGet)
	router.HandleFunc("/posts/period", publicposts.BuildPeriodIndexHandler(db, renderer)).Methods(http.MethodGet)
	router.HandleFunc("/posts/search", publicposts.BuildSearchHandler(db, renderer)).Methods(http.MethodGet)
	router.HandleFunc(`/posts/{date:\d{4}-\d{2}-\d{2}}{.*}`, publicposts.BuildLegacyPostRedirect()).Methods(http.MethodGet)
	router.HandleFunc(`/photos/{date:\d{4}-\d{2}-\d{2}}{.*}`,
		publicposts.BuildLegacyPostRedirect()).Methods(http.MethodGet)
	router.HandleFunc(`/archive/{month:\d{2}}-{day:\d{2}}`,
		publicposts.BuildLegacyPeriodRedirect()).Methods(http.MethodGet)
	router.HandleFunc("/posts/on-this-day/{month}-{day}",
		publicposts.BuildOnThisDayHandler(db, renderer)).Methods(http.MethodGet)
	router.HandleFunc("/posts/on-this-day", publicposts.BuildOnThisDayHandler(db, renderer)).Methods(http.MethodGet)
	router.HandleFunc("/posts/{postID}", publicposts.BuildGetHandler(db, renderer)).Methods(http.MethodGet)
	router.HandleFunc("/posts/{postID}/media", handlers.BuildMediaRedirectHelperHandler(db)).Methods(http.MethodGet)
	router.HandleFunc("/posts/", handlers.BuildRedirectHandler("/")).Methods(http.MethodGet)
	router.HandleFunc("/posts", handlers.BuildRedirectHandler("/")).Methods(http.MethodGet)

	router.HandleFunc("/locations",
		publiclocations.BuildIndexHandler(db, mapServerAPIKey, rendererMap)).
		Methods(http.MethodGet)
	router.HandleFunc("/locations/{locationID}", publiclocations.BuildGetHandler(db, renderer)).Methods(http.MethodGet)
	router.HandleFunc("/locations/{locationID}/map.jpg",
		publiclocations.BuildMapHandler(db, bucket, mapServerURL, mapServerAPIKey)).
		Methods(http.MethodGet)

	router.HandleFunc("/medias/{mediaID}/{file}.{kind}",
		publicmedias.BuildMediaHandler(db, bucket)).Methods(http.MethodGet)
	router.HandleFunc("/devices/{deviceID}/icon.{kind}",
		publicdevices.BuildIconHandler(db, bucket)).Methods(http.MethodGet)
	router.HandleFunc("/devices/{deviceID}", publicdevices.BuildShowHandler(db, renderer)).Methods(http.MethodGet)
	router.HandleFunc("/devices", publicdevices.BuildIndexHandler(db, renderer)).Methods(http.MethodGet)
	router.HandleFunc("/lenses/{lensID}.png", publicLenses.BuildIconHandler(db, bucket)).Methods(http.MethodGet)
	router.HandleFunc("/lenses/{lensID}", publicLenses.BuildShowHandler(db, renderer)).Methods(http.MethodGet)
	router.HandleFunc("/lenses", publicLenses.BuildIndexHandler(db, renderer)).Methods(http.MethodGet)

	router.HandleFunc("/random", publicposts.BuildRandomHandler(db)).Methods(http.MethodGet)

	router.HandleFunc("/trips/{tripID}", publictrips.BuildGetHandler(db, renderer)).Methods(http.MethodGet)

	adminRouter := router.PathPrefix(adminPath).Subrouter()

	// Apply email authentication middleware for non-development environments
	if environment != "development" {
		adminRouter.Use(InitMiddlewareEmailAuth(permittedEmailSuffix))
	}

	adminRouter.HandleFunc("", admin.BuildAdminIndexHandler(rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/", handlers.BuildRedirectHandler("/admin")).Methods(http.MethodGet)

	adminRouter.HandleFunc("/devices", devices.BuildIndexHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/devices", devices.BuildCreateHandler(db, bucket, rendererAdmin)).Methods(http.MethodPost)
	adminRouter.HandleFunc("/devices/new", devices.BuildNewHandler(rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/devices/{deviceID}", devices.BuildGetHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/devices/{deviceID}",
		devices.BuildFormHandler(db, bucket, rendererAdmin)).Methods(http.MethodPost)

	adminRouter.HandleFunc("/lenses", lenses.BuildIndexHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/lenses", lenses.BuildCreateHandler(db, bucket, rendererAdmin)).Methods(http.MethodPost)
	adminRouter.HandleFunc("/lenses/new", lenses.BuildNewHandler(rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/lenses/{lensID}", lenses.BuildGetHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/lenses/{lensID}", lenses.BuildFormHandler(db, bucket, rendererAdmin)).Methods(http.MethodPost)

	adminRouter.HandleFunc("/tags", tags.BuildIndexHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/tags", tags.BuildCreateHandler(db, rendererAdmin)).Methods(http.MethodPost)
	adminRouter.HandleFunc("/tags/new", tags.BuildNewHandler(rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/tags/{tagName}", tags.BuildGetHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/tags/{tagName}", tags.BuildFormHandler(db, rendererAdmin)).Methods(http.MethodPost)

	adminRouter.HandleFunc("/locations", locations.BuildIndexHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/locations", locations.BuildCreateHandler(db, rendererAdmin)).Methods(http.MethodPost)
	adminRouter.HandleFunc("/locations/new", locations.BuildNewHandler(rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/locations/select", locations.BuildSelectHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/locations/lookup",
		locations.BuildLookupHandler(mapServerAPIKey, rendererAdmin)).
		Methods(http.MethodGet)
	adminRouter.HandleFunc("/locations/{locationID}", locations.BuildGetHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/locations/{locationID}",
		locations.BuildFormHandler(db, bucket, rendererAdmin)).
		Methods(http.MethodPost)

	adminRouter.HandleFunc("/medias", medias.BuildIndexHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/medias", medias.BuildCreateHandler(db, bucket, rendererAdmin)).Methods(http.MethodPost)
	adminRouter.HandleFunc("/medias/new", medias.BuildNewHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/medias/{mediaID}", medias.BuildGetHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/medias/{mediaID}",
		medias.BuildFormHandler(db, bucket, rendererAdmin)).Methods(http.MethodPost)

	adminRouter.HandleFunc("/posts", posts.BuildIndexHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/posts", posts.BuildCreateHandler(db, rendererAdmin)).Methods(http.MethodPost)
	adminRouter.HandleFunc("/posts/new", posts.BuildNewHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/posts/{postID}", posts.BuildGetHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/posts/{postID}", posts.BuildFormHandler(db, rendererAdmin)).Methods(http.MethodPost)

	adminRouter.HandleFunc("/trips", trips.BuildIndexHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/trips", trips.BuildCreateHandler(db, rendererAdmin)).Methods(http.MethodPost)
	adminRouter.HandleFunc("/trips/new", trips.BuildNewHandler(rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/trips/{tripID}", trips.BuildGetHandler(db, rendererAdmin)).Methods(http.MethodGet)
	adminRouter.HandleFunc("/trips/{tripID}", trips.BuildFormHandler(db, rendererAdmin)).Methods(http.MethodPost)

	// catch all handlers to serve static files
	router.HandleFunc("/{.*}", buildStaticHandler()).Methods(http.MethodGet)

	return nil
}

func Serve(
	environment, hostname, addr, port string,
	db *sql.DB,
	bucket *blob.Bucket,
	mapServerURL, mapServerAPIKey string,
) {
	router := mux.NewRouter()
	router.Use(InitMiddlewareHTTPS(hostname, environment))

	// Email authentication configuration for reverse proxy
	permittedEmailSuffix := viper.GetString("admin.auth.permitted_email_suffix")

	err := Attach(
		router,
		db,
		bucket,
		mapServerURL,
		mapServerAPIKey,
		"/admin",
		environment,
		permittedEmailSuffix,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Check if port is available before starting server
	serverAddr := fmt.Sprintf("%s:%s", addr, port)
	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to bind to address %s: %v", serverAddr, err)
	}
	listener.Close()

	srv := &http.Server{
		Handler:      router,
		Addr:         serverAddr,
		WriteTimeout: 30 * time.Second,
		// this is set to 3 mins to allow many images to resized in a queue
		ReadTimeout: 180 * time.Second,
	}

	log.Printf("starting server on http://localhost:%s", port)
	log.Fatal(srv.ListenAndServe())
}
