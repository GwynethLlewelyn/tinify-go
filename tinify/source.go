package Tinify

import (
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
	ResizeMethodThumb = "thumb"	// new method!
)

type ResizeMethod string

// JSONified type for selecting resize options.
type ResizeOption struct {
	Method ResizeMethod `json:"method"`
	Width  int64        `json:"width"`
	Height int64        `json:"height"`
}

type Source struct {
	url      string
	commands map[string]interface{}
}

func newSource(url string, commands map[string]interface{}) *Source {
	s := new(Source)
	s.url = url
	if commands != nil {
		s.commands = commands
	} else {
		s.commands = make(map[string]interface{})
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

	body := map[string]interface{}{
		"source": map[string]interface{}{
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

func getSourceFromResponse(response *http.Response) (s *Source, err error) {
	location := response.Header["Location"]
	url := ""
	if len(location) > 0 {
		url = location[0]
	}

	s = newSource(url, nil)
	return
}

func (s *Source) ToFile(path string) error {
	result, err := s.toResult()
	if err != nil {
		return err
	}

	return result.ToFile(path)
}

// ToBuffer() extracts the raw data (an image) from the result.
// It's similar in concept to ToFile, but allows sending the data to STDOUT, for instance.
// (gwyneth 20230209)
func (s *Source) ToBuffer() (rawData []byte, err error) {
	result, err := s.toResult()
	if err != nil {
		return
	}

	rawData = result.Data()	// this is result.data, but may not be in the future, who knows? (gwyneth 20231209)
	if len(rawData) == 0 {
		err = fmt.Errorf("result returned zero bytes")
	}
	return
}


func (s *Source) Resize(option *ResizeOption) error {
	if option == nil {
		return errors.New("option for resize is required")
	}

	s.commands["resize"] = option

	return nil
}

var convertMIMETypes = map[string]string{
	"png":	"image/png",
	"jpeg":	"image/jpeg",
	"webp":	"image/webp",
}

// Extra type struct for JSONification purposes...
type ConvertOptions struct {
	Type string	`json:"type"`	// can be image/png, etc.
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
		allOpts += convertMIMETypes[option]
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
	Background string `json:"background"`	// "white", "black", or a hex colour.
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

func (s *Source) toResult() (r *Result, err error) {
	if len(s.url) == 0 {
		err = errors.New("url is empty")
		return
	}

	response, err := GetClient().Request(http.MethodGet, s.url, s.commands)
	if err != nil {
		return
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	r = NewResult(response.Header, data)
	return
}
