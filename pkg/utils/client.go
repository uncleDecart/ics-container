package utils

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

func DoHTTPRequest(reqType string, URL url.URL, body io.Reader, hdrs string) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest(reqType, URL.String(), body)

	if err != nil {
		return nil, err
	}

	if hdrs != "" {
		lines := strings.SplitN(hdrs, "\r\n", -1)
		for _, line := range lines {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				key := parts[0]
				val := parts[1]

				req.Header.Add(key, val)
			}
		}
	}

	return client.Do(req)
}
