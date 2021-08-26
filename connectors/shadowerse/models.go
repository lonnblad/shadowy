package shadowerse

import (
	"encoding/json"
	"net/http"
	"net/url"
)

func CreateShadeRequest(req *http.Request, body []byte) ShadeRequest {
	return ShadeRequest{
		URL:    *req.URL,
		Method: req.Method,
		Header: req.Header,
		Body:   body,
	}
}

func DecodeShadeRequest(req *http.Request) (shade ShadeRequest, err error) {
	err = readResponseData(req, &shade)
	return
}

type ShadeRequest struct {
	URL    url.URL
	Method string
	Header http.Header
	Body   []byte
}

func CreateShadeResponse(resp *http.Response, body []byte) ShadeResponse {
	return ShadeResponse{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       body,
	}
}

func DecodeShadeResponse(req *http.Request) (shade ShadeResponse, err error) {
	err = readResponseData(req, &shade)
	return
}

type ShadeResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func readResponseData(req *http.Request, v interface{}) (err error) {
	defer req.Body.Close()

	return json.NewDecoder(req.Body).Decode(v)
}
