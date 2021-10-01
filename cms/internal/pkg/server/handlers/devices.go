package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
)

func BuildIndexHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		results, err := database.AllDevices(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		response := ""
		for _, v := range results {
			response = fmt.Sprintf("%s %s", response, v.Name)
		}
		fmt.Fprintf(w, response)
	}
}
