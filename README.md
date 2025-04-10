# HTTPTestClient
HTTP Client without Test Server and Network Call

## Description
This is a simple HTTP client for testing purposes.
The client is designed to access the `http.Handler` directly, without the need for setting up a test server or listener.

This provides benefits such as:
- Preserving the call stack from the HTTP client call. (e.g. `http.Get`)
- Providing standard cookie jar for testing a series of requests.
- Injecting internal context into the request, for easy testing.
- Returning standard `*http.Client` to integrate with other libraries.

## How to use

### Standard Usage
Most test cases can use the client returned by `httptestclient.New()`.

```golang
package test

import (
  "testing"

  "github.com/khwong-c/httptestclient"
)

func TestMyHandler(t *testing.T) {
  handler := CreateMyServer() // e.g. http/chi/gorilla Mux

  // Create a new test client
  client := httptestclient.New(handler)

  // Make a request as usual
  resp, err := client.Get("http://localhost/my-handler")

  // ...
}

```

### Memorizing Cookies
Calling `httptestclient.NewWithCookieJar()` will create a client with cookie jar.

```golang
package test

import (
  "testing"

  "github.com/khwong-c/httptestclient"
)

func TestLogin(t *testing.T) {
  handler := CreateMyServer()

  // Create a new test client
  client := httptestclient.NewWithCookieJar(handler)

  // Make a sequence of requests with the same client 
  resp1, err := client.Get("http://localhost/login")
  resp2, err := client.Get("http://localhost/session")

  // ...
}

```

### Known Issues
- Test Client with Cookie Jar will not propagate the cookies, if the URL is not start with `http://` or `https://`. A stub host shall be provided.
  - Client can still make requests without this prefix, but the cookies will not be stored.
  - e.g. `/set-cookie` would not work; `http://localhost/set-cookie` will work.
  