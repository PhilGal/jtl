package rest

import (
	"net/http"
	"time"
)

//HTTPClient is a default HTTP client, a proxy over http.Client
var HTTPClient Client

//Client represents and interface for http.Client. Its main purpose is testing.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

func init() {
	HTTPClient = &http.Client{Timeout: time.Second * 30}
}
