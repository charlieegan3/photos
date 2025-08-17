package shared

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func ParseIDFromPath(r *http.Request, paramName string) (int64, error) {
	rawID, ok := mux.Vars(r)[paramName]
	if !ok {
		return 0, fmt.Errorf("%s is required", paramName)
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s was not integer", paramName)
	}

	return id, nil
}

func WriteError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(message))
}

func ValidateContentType(r *http.Request, expected string) error {
	contentType, ok := r.Header["Content-Type"]
	if !ok {
		return errors.New("Content-Type must be set")
	}

	if len(contentType) == 0 || contentType[0] != expected {
		return fmt.Errorf("Content-Type must be %s", expected)
	}

	return nil
}
