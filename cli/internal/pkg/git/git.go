package git

import (
	"log"
	"os"
)

var repoURL string
var username string
var accessToken string

func init() {
	repoURL = os.Getenv("GIT_REPO_URL")
	username = os.Getenv("GIT_USERNAME")
	accessToken = os.Getenv("GIT_ACCESS_TOKEN")

	if username == "" || repoURL == "" || accessToken == "" {
		log.Fatal("GIT_REPO_URL, GIT_USERNAME, and GIT_ACCESS_TOKEN must be set")
		os.Exit(1)
	}
}
