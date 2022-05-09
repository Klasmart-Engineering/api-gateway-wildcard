package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luraproject/lura/v2/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
}

func (m *Mock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

func (m *Mock) Header() http.Header {
	return map[string][]string{}
}

func (m *Mock) Write([]byte) (int, error) {
	return 0, nil
}

func (m *Mock) WriteHeader(statusCode int) {}

func TestServerPluginRequestModification(t *testing.T) {
	HandlerRegisterer.RegisterLogger(logging.NoOp)
	req := httptest.NewRequest("GET", "/foo/who", nil)
	mock := new(Mock)
	endpoints := map[string]string{"foo": "/__wildcard/bar"}
	mock.On("ServeHTTP", mock, req)

	forwardWildcardRequestToKrakendClient(mock, req, mock, endpoints)
	assert.Equal(t, "/who", req.Header.Get(headerName))
	assert.Equal(t, "/__wildcard/bar", req.URL.Path)
	mock.AssertCalled(t, "ServeHTTP", mock, req)
}
