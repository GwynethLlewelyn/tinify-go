package Tinify

import (
	"net/http"
	"os"
	"path/filepath"
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
	// Fix: by default, it was writing with permissions 0777
	err = os.WriteFile(path, r.data, os.FileMode(int(0644)))
	return err
}

// Retrieves the size of the image file, as described in the header.
// Note that some web server implementations might not return this value, or it might
// be incorrectly calculated.
func (r *Result) Size() int64 {
	return r.ResultMeta.size()
}

// Returns the MIME type of this object, as retrieved from the headers in the result.
func (r *Result) MediaType() string {
	return r.ResultMeta.mediaType()
}

// Deprecated: Alias to `MediatType()` for backwards compatibility.
func (r *Result) ContentType() string {
	return r.ResultMeta.mediaType()
}

// Returns the numbr of compressions made so far.
func (r *Result) CompressionCount() int64 {
	return r.ResultMeta.compressionCount()
}
