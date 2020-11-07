package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeVirtualUserReq(url, clientip string) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Real-Ip", clientip) // leverage the behavior of c.ClientIP()
	res, err := client.Do(req)
	return res, err
}

func Test_Basic(t *testing.T) {
	// =================================================================
	// Referencing from https://kpat.io/2019/06/testing-with-gin/
	// =================================================================

	// The setupServer method, that we previously refactored
	// is injected into a test server
	ts := httptest.NewServer(setupServer())
	// Shut down the server and block until all requests have gone through
	defer ts.Close()

	// Make a request to our server with the {base url}/ping
	resp, err := makeVirtualUserReq(ts.URL, "1.2.3.4")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %v", resp.StatusCode)
	}

	val, ok := resp.Header["Content-Type"]

	// Assert that the "content-type" header is actually set
	if !ok {
		t.Fatalf("Expected Content-Type header to be set")
	}

	// Assert that it was set as expected
	if val[0] != "application/json; charset=utf-8" {
		t.Fatalf("Expected \"application/json; charset=utf-8\", got %s", val[0])
	}
	defer resp.Body.Close()
	text, err := ioutil.ReadAll(resp.Body)
	t.Logf("%s \n", text)
}
