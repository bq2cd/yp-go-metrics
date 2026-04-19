package handlertest

import (
	"net/http"
	"testing"
)

// Response represents a subset of relevant information from HTTP response.
type Response struct {
	Status int
	Header http.Header
	Body   BodyData
}

// Requester is a wrapper over [http.Client] to perform HTTP requests.
type Requester struct {
	T      testing.TB
	client *http.Client
}

// NewRequester creates an instance of [Requester] with given [http.Client].
func NewRequester(t testing.TB, client *http.Client) *Requester {
	return &Requester{
		T:      t,
		client: client,
	}
}

// Do performs actual HTTP request based on provided method, url and [BodyData].
// The request's body will be compressed if `shouldCompress` is set to `true`.
// The method reads HTTP response's body and creates [Response] from it.
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
