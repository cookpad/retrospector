package mock

import (
	"io"
	"net/http"
)

type HTTPClient struct {
	Requests []*http.Request
	RespCode int
	RespBody io.ReadCloser
}

func (x *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	x.Requests = append(x.Requests, req)

	return &http.Response{
		StatusCode: x.RespCode,
		Body:       x.RespBody,
	}, nil
}
