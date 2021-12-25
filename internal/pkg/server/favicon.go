package server

import (
	"bytes"
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"io"
	"net/http"
)

//go:embed static/favicon.ico
var faviconContent []byte

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	faviconHash := sha1.New()
	faviconHash.Write(faviconContent)

	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Content-Type", "image/vnd.microsoft.icon")
	w.Header().Set("ETag", hex.EncodeToString(faviconHash.Sum(nil)))

	_, err := io.Copy(w, bytes.NewBuffer(faviconContent))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to return favicon content"))
		return
	}
}
