package tests

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/khwong-c/httptestclient"
)

func createHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello, World!"))
	})
	mux.HandleFunc("/error", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	})
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/fish", http.StatusFound)
	})
	mux.HandleFunc("/fish", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Fish"))
	})
	return mux
}

func createCookieHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/set-cookie", func(w http.ResponseWriter, _ *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "testcookie",
			Value: "testvalue",
		})
		_, _ = w.Write([]byte("Cookie Set"))
	})
	mux.HandleFunc("/get-cookie", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("testcookie")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("No Cookie Found"))
			return
		}
		cookieValue := cookie.Value
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Cookie Value: " + cookieValue))
	})
	return mux
}

type TestClientSuite struct {
	suite.Suite
	server           http.Handler
	serverWithCookie http.Handler
}

func TestClient(t *testing.T) {
	suite.Run(t, new(TestClientSuite))
}

func (s *TestClientSuite) SetupTest() {
	s.server = createHandler()
	s.serverWithCookie = createCookieHandler()
}

func (s *TestClientSuite) TestServeRequest() {
	client := httptestclient.New(s.server)
	for _, tc := range []struct {
		name         string
		endpoint     string
		expectedCode int
		expectedBody string
	}{
		{"GetRoot", "http://localhost/", http.StatusOK, "Hello, World!"},
		{"GetError", "http://localhost/error", http.StatusInternalServerError, "Internal Server Error"},
		{"FollowRedirect", "http://localhost/redirect", http.StatusOK, "Fish"},
	} {
		s.Run(tc.name, func() {
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tc.endpoint, nil)
			s.Require().NoError(err)
			if resp, err := client.Do(req); s.NoError(err) && s.NotNil(resp) {
				s.Equal(tc.expectedCode, resp.StatusCode)
				defer resp.Body.Close()
				if body, err := io.ReadAll(resp.Body); s.NoError(err) {
					s.Equal(tc.expectedBody, string(body))
				}
			}
		})
	}
}

func (s *TestClientSuite) TestCookieJar() {
	client, err := httptestclient.NewWithCookieJar(s.serverWithCookie)
	s.Require().NoError(err)

	// Make a request to set the cookie
	{
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost/set-cookie", nil)
		s.Require().NoError(err)
		resp, err := client.Do(req)
		s.Require().NoError(err)
		s.Equal(http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}

	// Make a request to get the cookie. Check if the cookie is set correctly.
	{
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost/get-cookie", nil)
		s.Require().NoError(err)
		resp, err := client.Do(req)
		s.Require().NoError(err)
		s.Equal(http.StatusOK, resp.StatusCode)
		defer resp.Body.Close()
		if resp, err := io.ReadAll(resp.Body); s.NoError(err) {
			s.Equal("Cookie Value: testvalue", string(resp))
		}
	}
}

// The Cookie Jar feature will not work if the URL is not a full URL.
func (s *TestClientSuite) TestCookieJar_FullURLIsRequired() {
	client, err := httptestclient.NewWithCookieJar(s.serverWithCookie)
	s.Require().NoError(err)

	// Make a request to set the cookie
	{
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/set-cookie", nil)
		s.Require().NoError(err)
		resp, err := client.Do(req)
		s.Require().NoError(err)
		s.Equal(http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}

	// Make a request to get the cookie. Check if the cookie is set correctly.
	{
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/get-cookie", nil)
		s.Require().NoError(err)
		resp, err := client.Do(req)
		s.Require().NoError(err)
		s.Equal(http.StatusBadRequest, resp.StatusCode)
		defer resp.Body.Close()
	}
}
