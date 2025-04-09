package tests

import (
	"fmt"

	"github.com/khwong-c/httptestclient"
)

//nolint:noctx
func ExampleNew() {
	// Build an HTTP Handler, without setting up a server and listener
	server := createHandler()
	// Create a test HTTP client
	client := httptestclient.New(server)
	// Make a request as usual
	resp, err := client.Get("/")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println("Response status:", resp.Status)
	// Output:
	// Response status: 200 OK
}
