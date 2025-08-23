package Tinify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Standard Tinify API endpoint.
const API_ENDPOINT = "https://api.tinify.com"

// This allows any consumer of this package to be aware of any proxies used.
var tinifyProxyTransport *http.Transport

// Type for the TinyPNG API client.
type Client struct {
	options map[string]any // List of options to call the Tinify API.
	key     string         // TinyPNG API key.
	proxy   string         // Specific HTTP(S) proxy server for this client.
}

// Creates a new TinyPNG API client by allocating some memory for it.
func NewClient(key string) (c *Client, err error) {
	c = new(Client)
	c.key = key
	return
}

// HTTP(S) request which can either send raw bytes (for an image) and/or a JSON-formatted request.
func (c *Client) Request(method string, urlRequest string, body any) (response *http.Response, err error) {
	// NOTE: this should go through a bit more validation. We are deferring such
	// validation to the Go library functions that do the actual request.
	if !strings.HasPrefix(urlRequest, "https") { // shouldn't we check for uppercase as well? (gwyneth 20231111)
		urlRequest = API_ENDPOINT + urlRequest
	}
	// Deal with HTTP(S) proxy for this request.
	tinifyProxyTransport.Proxy = c.reconfigureProxyTransport("") // the parameter is possibly irrelevant

	httpClient := http.Client{
		Transport: tinifyProxyTransport,
	}

	req, err := http.NewRequest(method, urlRequest, nil)
	if err != nil {
		return nil, fmt.Errorf("request to %q using method %q failed; error was: %s", urlRequest, method, err)
	}

	// Clunky! But it works :-)
	// If the body is a raw binary image, send it raw via an ioReader.
	// Otherwise, the body will need to be sent as JSON (per API). So first we construct a JSONified
	// representation of the struct we've got; and *then* send the result via an ioReader.
	switch b := body.(type) {
	case []byte:
		if len(b /*body.([]byte)*/) > 0 {
			req.Body = io.NopCloser(bytes.NewReader(b /*body.([]byte)*/))
		}
	case map[string]any:
		if len(b /*body.(map[string]interface{})*/) > 0 {
			bodyJSON, err2 := json.Marshal(body)
			if err2 != nil {
				err = err2
				return
			}
			req.Body = io.NopCloser(bytes.NewReader(bodyJSON))
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
func (c *Client) reconfigureProxyTransport(proxyURL string) func(*http.Request) (*url.URL, error) {
	reqProxy := http.ProxyURL(nil) // set to no proxy first.
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

	// Third attempt: check if the passed proxyURL value is any good:
	if reqProxy == nil && len(proxyURL) > 0 {
		tempURL, err := url.Parse(proxyURL)
		if err != nil {
			log.Printf("proxyURL parameter passed for this client must be a valid URL; got %q which gives error: %s", proxyURL, err)
			return nil
		}
		reqProxy = http.ProxyURL(tempURL)
	}

	// Fourth attempt: fallback to environment variables instead
	if reqProxy == nil {
		reqProxy = http.ProxyFromEnvironment
	}
	// Note: if reqProxy is *still* `nil`, that's correct and appropriate for *no proxy*
	return reqProxy
}

// Tinify module initialisation.
// Currently just initialises tinifyProxyTransport as the default-
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
}
