package Tinify

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// Object returned by a call to the Tinify API.
// Note that the metadata is its own object and handled separately.
type Result struct {
	data        []byte // Raw image data.
	*ResultMeta        // Additional metadata returned by TinyPNG, namely, the file location generated.
}

// Constructor for the `Result` object.
func NewResult(meta http.Header, data []byte) *Result {
	r := new(Result)
	r.ResultMeta = NewResultMeta(meta) // also handle metadata.
	r.data = data
	return r
}

// Returns the raw body of the call.
func (r *Result) Data() []byte {
	return r.data
}

// Retrieves the actual body of the call, which may be the raw image data.
func (r *Result) ToBuffer() []byte {
	return r.Data()
}

// Writes this object (an image file) to disk.
func (r *Result) ToFile(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, r.data, os.ModePerm)
	return err
}

// Retrieves the size of the image file, as described in the header.
// Note that some web server implementations might not return this value, or it might
// be incorrectly calculated.
func (r *Result) Size() int64 {
	s := r.meta["Content-Length"]
	if len(s) == 0 {
		return 0
	}

	size, _ := strconv.Atoi(s[0]) // Atoi returns 0 if error
	return int64(size)
}

// Returns the MIME type of this object, as retrieved from the headers in the result.
func (r *Result) MediaType() string {
	arr := r.meta["Content-Type"]
	if len(arr) == 0 {
		return ""
	}
	return arr[0]
}

// Deprecated: alias to `ContentType()` for backwards compatibility.
func (r *Result) ContentType() string {
	return r.MediaType()
}
