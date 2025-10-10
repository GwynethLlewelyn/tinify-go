package Tinify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	ResizeMethodScale = "scale"
	ResizeMethodFit   = "fit"
	ResizeMethodCover = "cover"
	ResizeMethodThumb = "thumb" // new method!
)

type ResizeMethod string

// JSONified type for selecting resize options.
type ResizeOption struct {
	Method ResizeMethod `json:"method"`
	Width  int64        `json:"width,omitempty"`
	Height int64        `json:"height,omitempty"`
}

// Main object type for returning a result.
type Source struct {
	url              string         // URL to retrieve from.
	commands         map[string]any // Commands passed to the Tinify API.
	compressionCount string         // This is the number of compressions made with this API key this month; may become an integer in the future,
}

// JSONified type for error messages from the Tinify API, if present.
type ErrorMessage struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func newSource(url string, commands map[string]any) *Source {
	s := new(Source)
	s.url = url
	if commands != nil {
		s.commands = commands
	} else {
		s.commands = make(map[string]any)
	}

	return s
}

func FromFile(path string) (s *Source, err error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return
	}

	return FromBuffer(buf)
}

func FromBuffer(buf []byte) (s *Source, err error) {
	response, err := GetClient().Request(http.MethodPost, "/shrink", buf)
	if err != nil {
		return
	}

	s, err = getSourceFromResponse(response)
	return
}

func FromUrl(url string) (s *Source, err error) {
	if len(url) == 0 {
		err = errors.New("URL is required")
		return
	}

	body := map[string]any{
		"source": map[string]any{
			"url": url,
		},
	}

	response, err := GetClient().Request(http.MethodPost, "/shrink", body)
	if err != nil {
		return
	}

	s, err = getSourceFromResponse(response)
	return
}

// getSourceFromResponse tries to retrieve the URL that the Tinify API created to download the processed image.
func getSourceFromResponse(response *http.Response) (s *Source, err error) {
	location := response.Header["Location"]
	url := ""
	if len(location) > 0 && response.StatusCode != http.StatusBadRequest {
		url = location[0]
	} else {
		return nil, fmt.Errorf("empty location and/or status error %d", response.StatusCode)
	}

	s = newSource(url, nil)
	// Get number of compressions for this API key for this month, it comes in a header of its own.
	// If the request didn't have such a header, that's ok, it'll just be an empty sring.
	// (gwyneth 29250713)
	s.compressionCount = response.Header["Compression-Count"][0]
	return
}

// ToFile is a wrapper that grabs the content of the result and writes to a file.
// The compression count is discarded.
//
// Obsolete: kept here only for compatibility purposes.
func (s *Source) ToFile(path string) (err error) {
	_, err = s.ToFileC(path)
	return
}

// ToFileC is a wrapper that grabs the content of the result and writes to a file.
// The compression count is returned as well.
//
// Supersedes `ToFile()`.
func (s *Source) ToFileC(path string) (int64, error) {
	result, err := s.toResult()
	if err != nil {
		return result.compressionCount(), err
	}

	return result.compressionCount(), result.ToFile(path)
}

// ToBuffer extracts the raw data (an image) from the result.
// It's similar in concept to ToFile, but allows sending the data to STDOUT, for instance.
// (gwyneth 20230209)//
//
// Obsolete: kept here only for compatibility purposes.
func (s *Source) ToBuffer() (rawData []byte, err error) {
	rawData, _, err = s.ToBufferC()
	return
}

// ToBufferC extracts the raw data (an image) from the result, also returning
// the compression count.
//
// Supersedes `ToBuffer()`
func (s *Source) ToBufferC() (rawData []byte, count int64, err error) {
	result, err := s.toResult()
	if err != nil {
		return
	}

	// Extract the compression count, even if the subsequent raw data
	// extraction fails.
	count = result.compressionCount()

	rawData = result.Data() // this is result.data, but may not be in the future, who knows? (gwyneth 20231209)
	if len(rawData) == 0 {
		err = fmt.Errorf("result returned zero bytes")
	}
	return
}

// Checks errors in the list of commands for a resizing operation.
func (s *Source) Resize(option *ResizeOption) error {
	if option == nil {
		return errors.New("option for resize is required")
	}
	// "scale" can only have width or height set, but not both!
	if option.Method == ResizeMethodScale {
		if option.Width != 0 && option.Height != 0 {
			return errors.New("resize with scale method can only have either width or height set, but not both")
		}
		if option.Width == 0 && option.Height == 0 {
			return errors.New("resize with scale method cannot have width and height both set to zero")
		}
	} else {
		// for all other methods, the smallest possible value is 1!
		if option.Width < 1 {
			return errors.New("width must be >=1")
		}
		if option.Height < 1 {
			return errors.New("height must be >=1")
		}
	}
	s.commands["resize"] = option

	return nil
}

var ConvertMIMETypes = map[string]string{
	"png":  "image/png",
	"jpeg": "image/jpeg",
	"webp": "image/webp",
	"avif": "image/avif",
}

// Extra type struct for JSONification purposes...
type ConvertOptions struct {
	Type string `json:"type"` // can be image/png, etc.
}

// Converts the image to one of several possible choices, returning the smallest.
func (s *Source) Convert(options []string) error {
	if len(options) == 0 {
		return errors.New("at least one option for convert is required")
	}
	// quick & dirty
	allOpts := ""
	for i, option := range options {
		if i != 0 {
			allOpts += ","
		}
		allOpts += ConvertMIMETypes[option]
	}
	// Should never happen...
	if len(allOpts) == 0 {
		return errors.New("concatenation of MIME types unexpectedly failed")
	}
	// Allocate some memory for the convert options, one never knows...
	convertOptions := new(ConvertOptions)
	convertOptions.Type = allOpts
	s.commands["convert"] = convertOptions

	return nil
}

// JSONified type for transform options, currently only "background" is supported.
type TransformOptions struct {
	Background string `json:"background"` // "white", "black", or a hex colour.
}

// Transforms the transparency colour into the desired background colour.
// Valid options are "white", "black", or a hex colour.
func (s *Source) Transform(option *TransformOptions) error {
	if option == nil {
		return errors.New("at least one option for transform is required")
	}

	s.commands["transform"] = option

	return nil
}

// toResult does the actual remote API call. It returns either a *Result or nil with an error
// message covering most possibilities of failure.
// The Tinify API specifies that all errors come as properly-formatted JSON, but we check even for that.
func (s *Source) toResult() (r *Result, err error) {
	if len(s.url) == 0 {
		err = errors.New("url is empty")
		return
	}

	response, err := GetClient().Request(http.MethodGet, s.url, s.commands)
	if err != nil {
		return
	}

	// NOTE: if the request succeeds, but the API found an error, it returns with a JSON
	// indicating the error.
	data, err := io.ReadAll(response.Body)

	// did we get an error code from the API call?
	// Note: we consider all JSON answers as "errors", evn if the API doesn't mandate that.
	if response.StatusCode >= 400 || response.Header.Get("Content-Type") == "application/json" {
		// we got an error but couldn't retrieve any data:
		if err != nil {
			// we can only retrieve the Status line with a short error
			return nil, errors.New(response.Status)
		}
		// otherwise, we can return the error by unmarshalling the received JSONified error:
		var errMsg ErrorMessage

		if jErr := json.Unmarshal(data, &errMsg); jErr != nil {
			// Unmarshalling failed, but we still have
			return nil, fmt.Errorf("Tinify API call failed, HTTP status was %q, couldn't unmarshal JSON body: error %q", response.Status, jErr)
		}
		// We've successfully decoded the error message, so we can return it:
		return nil, fmt.Errorf("Tinify API call failed, HTTP status was %q. Error: %s Message: %s",
			response.Status, errMsg.Error, errMsg.Message)
	}
	// At this stage, the only error we have is from a failed decoded body data.
	if err != nil {
		return
	}

	// No errors found. The result can be sent back to the caller.
	r = NewResult(response.Header, data)
	return
}
