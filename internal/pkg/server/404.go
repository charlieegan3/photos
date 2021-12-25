package server

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func notFound(w http.ResponseWriter, r *http.Request) {
	logrus.NewEntry(logrus.New()).WithFields(logrus.Fields{
		"status": http.StatusNotFound,
		"path":   r.URL.Path,
		"method": r.Method,
	}).Info("not found")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not found"))
}
