package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

	var Tests = []struct {
		testName            string
		method              string
		inputUrl            string
		endpoints           map[string]string
		shouldHaveRerouted  bool
		shouldCallHttpServe bool
		targetUrl           string
	}{
		{
			testName:            "valid get scenario",
			method:              "GET",
			inputUrl:            "/foo/who",
			endpoints:           map[string]string{"foo": "/__wildcard/bar"},
			shouldHaveRerouted:  true,
			shouldCallHttpServe: true,
			targetUrl:           "/who",
		},
		{
			testName:            "valid post scenario",
			method:              "POST",
			inputUrl:            "/foo/who",
			endpoints:           map[string]string{"foo": "/__wildcard/bar"},
			shouldHaveRerouted:  true,
			shouldCallHttpServe: true,
			targetUrl:           "/who",
		},
		{
			testName:            "valid scenario with multiple endpoints",
			method:              "GET",
			inputUrl:            "/bar/hello",
			endpoints:           map[string]string{"foo": "/__wildcard/bar", "bar": "/__wildcard/foo"},
			shouldHaveRerouted:  true,
			shouldCallHttpServe: true,
			targetUrl:           "/hello",
		},
		{
			testName:            "no matches",
			method:              "GET",
			inputUrl:            "/bar/who",
			endpoints:           map[string]string{},
			shouldHaveRerouted:  false,
			shouldCallHttpServe: true,
			targetUrl:           "",
		},
		{
			testName:            "root url",
			method:              "GET",
			inputUrl:            "/",
			endpoints:           map[string]string{"foo": "/__wildcard/bar"},
			shouldHaveRerouted:  false,
			shouldCallHttpServe: true,
			targetUrl:           "",
		},
		{
			testName:            "directly trying to hit the wildcard url",
			method:              "GET",
			inputUrl:            "/__wildcard/foo",
			endpoints:           map[string]string{"foo": "/__wildcard/foo"},
			shouldHaveRerouted:  false,
			shouldCallHttpServe: false,
			targetUrl:           "",
		},
	}

	HandlerRegisterer.RegisterLogger(logging.NoOp)
	for _, tt := range Tests {
		req := httptest.NewRequest(tt.method, tt.inputUrl, nil)
		mock := new(Mock)
		mock.On("ServeHTTP", mock, req)

		prefixPath := strings.Split(tt.inputUrl, "/")[1]
		targetInternalUrl := tt.endpoints[prefixPath]

		forwardWildcardRequestToKrakendClient(mock, req, mock, tt.endpoints)

		if tt.shouldHaveRerouted {
			assert.Equal(t, tt.targetUrl, req.Header.Get(headerName))
			assert.Equal(t, targetInternalUrl, req.URL.Path)
		} else {
			assert.Equal(t, tt.inputUrl, req.URL.Path)
		}

		if tt.shouldCallHttpServe {
			mock.AssertCalled(t, "ServeHTTP", mock, req)
		} else {
			mock.AssertNotCalled(t, "ServeHTTP")
			mock.AssertNotCalled(t, "ServeHTTP", mock, req)
		}
	}
}
