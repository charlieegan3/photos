package proxy

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
)

var proxyURL *url.URL
var proxyToken string

func init() {
	rawURL := os.Getenv("PROXY_URL")
	if rawURL == "" {
		log.Fatal("PROXY_URL is not set")
		os.Exit(1)
	}

	var err error
	proxyURL, err = url.Parse(rawURL)
	if err != nil {
		log.Fatalf("failed to parse proxy url (%s): %s", rawURL, err)
		os.Exit(1)
	}

	proxyToken = os.Getenv("PROXY_TOKEN")
	if proxyToken == "" {
		log.Fatal("PROXY_TOKEN is not set")
		os.Exit(1)
	}
}

// GetURLViaProxy will make a get request using a simple-proxy endpoint
// and forward the result
func GetURLViaProxy(requestURL string, headers map[string]string) (int, []byte, error) {
	fmt.Println(requestURL)
	response := &http.Response{}

	q := proxyURL.Query()
	q.Add("url", requestURL)
	proxyURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", proxyURL.String(), nil)
	if err != nil {
		return 0, []byte{}, errors.Wrap(err, "failed to create proxy request")
	}

	client := &http.Client{}

	req.Header.Add("Authorization", "bearer "+proxyToken)

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	response, err = client.Do(req)
	if err != nil {
		return 0, []byte{}, errors.Wrap(err, "failed to get via proxy")
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	return response.StatusCode, body, nil
}
