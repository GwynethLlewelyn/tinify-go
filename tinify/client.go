package Tinify

import (
	"bytes"
	"fmt"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const API_ENDPOINT = "https://api.tinify.com"

var tinifyProxyTransport *http.Transport
// func(*http.Request) (*url.URL, error)

// Type for the TinyPNG API client.
type Client struct {
	options map[string]interface{}
	key     string	// TinyPNG API key.
	proxy	string	// Specific HTTP(S) proxy server for this client.
}

// Creates a new TinyPNG API client by allocating some memory for it.
func NewClient(key string) (c *Client, err error) {
	c = new(Client)
	c.key = key
	return
}

// HTTP(S) request which can either send raw bytes (for an image) and/or a JSON-formatted request.
func (c *Client) Request(method string, urlRequest string, body interface{}) (response *http.Response, err error) {
	// NOTE: this should go through a bit more validation. We are deferring such
	// validation to the Go library functions that do the actual request.
	if !strings.HasPrefix(urlRequest, "https") {	// shouldn't we check for uppercase as well? (gwyneth 20231111)
		urlRequest = API_ENDPOINT + urlRequest
	}
	// Deal with HTTP(S) proxy for this request.
	// var error err
	tinifyProxyTransport.Proxy = c.reconfigureProxyTransport(urlRequest)

	httpClient := http.Client{
		Transport: tinifyProxyTransport,
	}

	req, err := http.NewRequest(method, urlRequest, nil)
	if err != nil {
		return nil, fmt.Errorf("request to %q using method %q failed; error was: %s", urlRequest, method, err)
	}

	// Clunky! But it works :-)
	switch b := body.(type) {
	case []byte:
		if len(b/*body.([]byte)*/) > 0 {
			req.Body = io.NopCloser(bytes.NewReader(b/*body.([]byte)*/))
		}
	case map[string]interface{}:
		if len(b/*body.(map[string]interface{})*/) > 0 {
			body2, err2 := json.Marshal(body)
			if err2 != nil {
				err = err2
				return
			}
			req.Body = io.NopCloser(bytes.NewReader(body2))
		}
		req.Header["Content-Type"] = []string{"application/json"}
	default:
		return nil, fmt.Errorf("invalid request body; must be either an image or a JSON object")
	}

	req.SetBasicAuth("api", c.key)

	response, err = httpClient.Do(req)
	return
}

// Attempts to reconfigure an _existing_ Transport with a proxy.
func (c *Client) reconfigureProxyTransport(proxyURL string) (func(*http.Request) (*url.URL, error)) {
	reqProxy := http.ProxyURL(nil)	// set to no proxy first.
	// check if our global variable has been set:
	if len(proxy) > 0 {
		tempURL, err := url.Parse(proxy)
		if err != nil {
			log.Printf("global proxy must be a valid URL; got %q which gives error: %s\n", proxy, err)
			return nil
		}
		reqProxy = http.ProxyURL(tempURL)
	}
	// Second attempt: override it with the proxy setting for _this_ client instead:
	if reqProxy == nil && len(c.proxy) > 0 {
		tempURL, err := url.Parse(c.proxy)
		if err != nil {
			log.Printf("proxy set for this client must be a valid URL; got %q which gives error: %s", c.proxy, err)
			return nil
		}
		reqProxy = http.ProxyURL(tempURL)
	}
	// Third attempt: fallback to environment variables instead
	if reqProxy == nil {
		reqProxy = http.ProxyFromEnvironment
	}

	return reqProxy
}

// Tinify module initialisation.
// Currently just initialises tinifyProxyTransport as the equivalent
func init() {
	// initialise the transport; instructions say that transports should be reused, not
	// created on demand; by default, uses whatever proxies are defined on the environment.
	tinifyProxyTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		// DialContext: defaultTransportDialContext(&net.Dialer{
		// 	Timeout:   30 * time.Second,
		// 	KeepAlive: 30 * time.Second,
		// }),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	// .Clone(http.DefaultTransport)
/* 	 &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	} */
}
