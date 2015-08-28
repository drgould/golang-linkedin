package linkedin

import (
	"errors"
	"net/http"
)

// Responders are callbacks that receive and http request and return a mocked response.
type Responder func(*http.Request) (*http.Response, error)

// NoResponderFound is returned when no responders are found for a given HTTP method and URL.
var NoResponderFound = errors.New("no responder found")

// MockTransport implements http.RoundTripper, which fulfills single http requests issued by
// an http.Client.  This implementation doesn't actually make the call, instead defering to
// the registered list of responders.
type MockTransport struct {
	FailNoResponder bool
	responders      map[string]Responder
}

// RoundTrip is required to implement http.MockTransport.  Instead of fulfilling the given request,
// the internal list of responders is consulted to handle the request.  If no responder is found
// an error is returned, which is the equivalent of a network error.
func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	u := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path + "?" + req.URL.Query().Encode()

	key := req.Method + " " + u

	// scan through the responders and find one that matches our key
	for k, r := range m.responders {
		if k != key {
			continue
		}
		return r(req)
	}

	// if we've been told to error when no match was found
	if m.FailNoResponder {
		return nil, NoResponderFound
	}

	// fallback to the default roundtripper
	return http.DefaultTransport.RoundTrip(req)
}

// RegisterResponder adds a new responder, associated with a given HTTP method and URL.  When a
// request comes in that matches, the responder will be called and the response returned to the client.
func (m *MockTransport) RegisterResponder(method, url string, responder Responder) {
	if m.responders == nil {
		m.responders = make(map[string]Responder)
	}
	m.responders[method+" "+url] = responder
}

// DefaultMockTransport allows users to easily and globally alter the default RoundTripper for
// all http requests.
var DefaultMockTransport = &MockTransport{}

// Activate replaces the `Transport` on the `http.DefaultClient` with our `DefaultMockTransport`.
func Activate(failNoResponder bool) {
	DefaultMockTransport.FailNoResponder = failNoResponder
	http.DefaultClient.Transport = DefaultMockTransport
}

// Deactivate replaces our `DefaultMockTransport` with the `http.DefaultTransport`.
func Deactivate() {
	http.DefaultClient.Transport = http.DefaultTransport
}

// RegisterResponder adds a responder to the `DefaultMockTransport` responder table.
func RegisterResponder(method, url string, responder Responder) {
	DefaultMockTransport.RegisterResponder(method, url, responder)
}
