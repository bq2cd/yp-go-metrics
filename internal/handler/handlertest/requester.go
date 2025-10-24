package handlertest

import (
	"net/http"
	"testing"
)

type Response struct {
	Status int
	Header http.Header
	Body   BodyData
}

type Requester struct {
	T      testing.TB
	client *http.Client
}

func NewRequester(t testing.TB, client *http.Client) *Requester {
	return &Requester{
		T:      t,
		client: client,
	}
}

func (rqr *Requester) Do(method, url string, body BodyData, shouldCompress bool) (*Response, error) {
	req := body.NewRequest(method, url, shouldCompress)
	resp, err := rqr.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	return &Response{
		Status: resp.StatusCode,
		Header: resp.Header,
		Body:   NewBodyDataFromResponse(rqr.T, resp),
	}, nil
}
