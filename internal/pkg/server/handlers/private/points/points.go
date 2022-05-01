package points

import (
	"database/sql"
	"encoding/json"
	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func BuildCreateHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		if val, ok := r.Header["Content-Type"]; !ok || val[0] != "application/json" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Content-Type must be 'multipart/form-data'"))
			return
		}

		upload := struct {
			Latitude         float64 `json:"lat"`
			Longitude        float64 `json:"lon"`
			Accuracy         float64 `json:"acc"`
			VerticalAccuracy float64 `json:"vac"`
			Velocity         float64 `json:"vel"`
			Altitude         float64 `json:"alt"`
			Connection       string  `json:"conn"`
			Topic            string  `json:"topic"`
			Time             int64   `json:"tst"`
		}{}

		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		err = json.Unmarshal(bytes, &upload)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		points := []models.Point{
			{
				Latitude:         upload.Latitude,
				Longitude:        upload.Longitude,
				Velocity:         upload.VerticalAccuracy,
				Altitude:         upload.Altitude,
				Accuracy:         upload.Accuracy,
				VerticalAccuracy: upload.VerticalAccuracy,
				WasOffline:       upload.Connection == "o",
				CreatedAt:        time.Unix(upload.Time, 0),
			},
		}

		topicParts := strings.Split(upload.Topic, "/")
		if len(topicParts) != 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected topic format"))
			return
		}

		_, err = database.CreatePoints(db, "owntracks/handler", topicParts[2], "owntracks/event", nil, points)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}
