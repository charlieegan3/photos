package server

import (
	"crypto/sha1"
	"embed"
	"encoding/hex"
	"fmt"
	"net/http"
)

//go:embed static/css/*
var cssContent embed.FS

func buildStylesHandler() (handler func(http.ResponseWriter, *http.Request), err error) {
	normalizeData, err := cssContent.ReadFile("static/css/normalize.min.css")
	if err != nil {
		return handler, fmt.Errorf("failed to load normalize css data: %w", err)
	}

	tachyonsData, err := cssContent.ReadFile("static/css/tachyons.min.css")
	if err != nil {
		return handler, fmt.Errorf("failed to load tachyons css data: %w", err)
	}

	siteStyleData, err := cssContent.ReadFile("static/css/styles.css")
	if err != nil {
		return handler, fmt.Errorf("failed to load site styles css data: %w", err)
	}

	allStyleData := ""
	for _, b := range []*[]byte{&normalizeData, &tachyonsData, &siteStyleData} {
		allStyleData += string(*b) + "\n"
	}

	styleHash := sha1.New()
	styleHash.Write([]byte(allStyleData))

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=60")
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("ETag", hex.EncodeToString(styleHash.Sum(nil)))

		fmt.Fprint(w, allStyleData)
	}, nil
}
