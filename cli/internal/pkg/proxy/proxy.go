package proxy

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
)

var proxyURLRaw string
var proxyToken string
var proxyEnable bool

func init() {
	proxyURLRaw = os.Getenv("PROXY_URL")
	if proxyURLRaw == "" {
		log.Fatal("PROXY_URL is not set")
		os.Exit(1)
	}

	var err error
	_, err = url.Parse(proxyURLRaw)
	if err != nil {
		log.Fatalf("failed to parse proxy url (%s): %s", proxyURLRaw, err)
		os.Exit(1)
	}

	proxyToken = os.Getenv("PROXY_TOKEN")
	if proxyToken == "" {
		log.Fatal("PROXY_TOKEN is not set")
		os.Exit(1)
	}

	proxyEnableStr := os.Getenv("PROXY_ENABLE")
	proxyEnable = true
	if proxyEnableStr == "false" {
		proxyEnable = false
	}
}

// GetURLViaProxy will make a get request using a simple-proxy endpoint
// and forward the result
func GetURLViaProxy(requestURL string, headers map[string]string) (int, []byte, error) {
	log.Printf("fetching via proxy: %s", requestURL)
	response := &http.Response{}

	var urlToRequest string
	if proxyEnable {
		// need a fresh copy of the url since we change the RawQuery each request
		proxyURL, err := url.Parse(proxyURLRaw)
		if err != nil {
			return 0, []byte{}, errors.Wrap(err, "failed to parse raw proxy url")
		}

		q := proxyURL.Query()
		q.Add("url", requestURL)
		proxyURL.RawQuery = q.Encode()

		urlToRequest = proxyURL.String()
	} else {
		urlToRequest = requestURL
	}

	req, err := http.NewRequest("GET", urlToRequest, nil)
	if err != nil {
		return 0, []byte{}, errors.Wrap(err, "failed to create proxy request")
	}

	client := &http.Client{}

	if proxyEnable {
		req.Header.Add("Authorization", "bearer "+proxyToken)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	response, err = client.Do(req)
	if err != nil {
		if proxyEnable {
			return 0, []byte{}, errors.Wrap(err, "failed to get via proxy")
		} else {
			return 0, []byte{}, errors.Wrap(err, "failed to get url")
		}
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	return response.StatusCode, body, nil
}
