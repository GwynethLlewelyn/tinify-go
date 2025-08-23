// Retrieves TinyPNG-specific header tags, converting them to the proper values
// as returned by the API. ResultMeta **must** be initialised with the headers
// received from the API call.
package Tinify

import (
	"net/http"

	"strconv"
)

type ResultMeta struct {
	meta http.Header
}

// NewResultMMeta creates a metadata object, reading the data
func NewResultMeta(meta http.Header) *ResultMeta {
	r := new(ResultMeta)
	r.meta = meta
	return r
}

func (r *ResultMeta) width() int64 {
	w := r.meta["Image-Width"]
	if len(w) == 0 {
		return 0
	}
	width, _ := strconv.Atoi(w[0])
	return int64(width)
}

func (r *ResultMeta) height() int64 {
	h := r.meta["Image-Height"]
	if len(h) == 0 {
		return 0
	}
	height, _ := strconv.Atoi(h[0])
	return int64(height)
}

func (r *ResultMeta) location() string {
	arr := r.meta["Location"]
	if len(arr) == 0 {
		return ""
	}
	return arr[0]
}

// TODO(gwyneth): add Compression-Count and other potentially relevant headers
// (gwyneth 20231003)

// compressionCount returns how many times the user has invoked API calls.
// The number is supposed to be reset every month, and there is a limit on the number of free calls
// per month. Some operations will 'consume' more than one invocation.
func (r *ResultMeta) compressionCount() int64 {
	c := r.meta["Compression-Count"]
	if len(c) == 0 {
		return 0
	}
	compC, _ := strconv.Atoi(c[0])
	return int64(compC)
}
