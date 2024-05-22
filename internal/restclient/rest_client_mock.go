package restclient

import (
	"context"
	"fmt"
)

// MockResponse is used in Unit Testing to mock expected REST responses.
// It validates that the request matches ExpectedMethod and ExpectedURL, to return the other elements.
type MockResponse struct {
	ExpectedMethod string
	ExpectedURL    string
	StatusCode     int
	Response       RestResponse
	Err            error
}

// NewMockedRestClient is used in Unit Testing to mock expected REST responses.
func NewMockedRestClient(responses []MockResponse) (*RestClient, error) {
	cxProfile := ConnectionProfile{
		Hostname: "",
		Username: "",
		Password: "",
	}
	newRestClient, err := NewClient(context.Background(), cxProfile, "resource/version", 600)
	if err != nil {
		panic(err)
	}
	newRestClient.mode = "mock"
	newRestClient.responses = responses

	return newRestClient, nil
}

func (r *RestClient) mockCallAPIMethod(method string, baseURL string, query *RestQuery, body map[string]any) (int, RestResponse, error) {
	if len(r.responses) == 0 {
		panic(fmt.Sprintf("Unexpected request: %s %s", method, baseURL))
	}
	expectedResponse := r.responses[0]
	if expectedResponse.ExpectedMethod != method || expectedResponse.ExpectedURL != baseURL {
		if len(r.responses) == 0 {
			panic(fmt.Sprintf("Unexpected request: %s %s, expecting %s %s", method, baseURL, expectedResponse.ExpectedMethod, expectedResponse.ExpectedURL))
		}
	}
	// remove element now that we know it is consumed
	r.responses = r.responses[1:]

	return expectedResponse.StatusCode, expectedResponse.Response, expectedResponse.Err
}
