package Tinify

import "errors"

const VERSION = "v0.1.0"	// using semantic versioning; 1.0 is considered "stable"...

var (
	key		string	// TinyPNG API Key, as obtained through https://tinypng.com/developers.
	client	*Client	//
	proxy	string	//
)

// This function allows Tinify.SetKey() to be valid Go code.
func SetKey(setKey string) {
	key = setKey
}

// Go will automatically use proxies, but that's fine.
func Proxy(setProxy string) {
	proxy = setProxy
}


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
