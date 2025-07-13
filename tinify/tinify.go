// Unofficial implementation of the Tinify API for image manipulation.
//
// Author:	"gwpp"
// Email:	"ganwenpeng1993@163.com",
package Tinify

import (
	"errors"
)

const VERSION = "v0.2.0" // using semantic versioning; 1.0 is considered "stable"...

var (
	key    string  // Tinify API Key, as obtained through https://tinypng.com/developers.
	client *Client // Default Tinify API client.
	proxy  string  // Proxy used just for the Tinify API.
)

// Sets the global Tinify API key for the module.
// NOTE: This function allows `Tinify.SetKey()` to be valid Go code.
func SetKey(setKey string) {
	key = setKey
}

// Go will automatically use proxies, but that's fine, we can still override them.
func Proxy(setProxy string) {
	proxy = setProxy
}

// Returns a new Client after checking that the stored API key is valid.
func GetClient() *Client {
	if len(key) == 0 {
		panic(errors.New("provide an API key with Tinify.SetKey(key string)"))
	}

	if client == nil {
		c, err := NewClient(key)
		if err != nil {
			panic(errors.New("provide an API key with Tinify.SetKey(key string)"))
		}
		client = c
	}
	return client
}
