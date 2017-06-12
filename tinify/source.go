package Tinify

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/astaxie/beego/logs"
)

type Source struct {
	url      string
	commands map[string]interface{}
}

func newSource(url string, commands map[string]interface{}) *Source {
	s := new(Source)
	s.url = url
	s.commands = commands
	return s
}

func FromFile(path string) (s *Source, err error) {
	buf, err := ioutil.ReadFile(path)
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
	logs.Info(url)
	if len(url) == 0 {
		err = errors.New("url is required")
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
	logs.Info("%+v", response.Header)
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

func (s *Source) toResult() (r *Result, err error) {
	if len(s.url) == 0 {
		err = errors.New("url is empty")
		return
	}

	body := make([]byte, 0)
	if len(s.commands) > 0 {
		body, err = json.Marshal(s.commands)
		if err != nil {
			return
		}
	}
	response, err := GetClient().Request(http.MethodGet, s.url, body)
	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	r = NewResult(response.Header, data)
	return
}