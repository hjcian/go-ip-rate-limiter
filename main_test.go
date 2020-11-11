package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func AssertErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func AssertStatus200(t *testing.T, resp *http.Response) {
	t.Helper()
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %v", resp.StatusCode)
	}
}

func AssertStatusError(t *testing.T, resp *http.Response) {
	t.Helper()
	if resp.StatusCode == 200 {
		t.Errorf("Expected status code != 200, got %v", resp.StatusCode)
	}
}

func AssertRemaining(t *testing.T, got, expect string) {
	t.Helper()
	if got != expect {
		t.Errorf("Expected Remaining %v, got %v", expect, got)
	}
}

func AssertUsed(t *testing.T, got, expect string) {
	t.Helper()
	if got != expect {
		t.Errorf("Expected Used %v, got %v", expect, got)
	}
}

func sendReq(url, clientip string) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Real-Ip", clientip) // leverage the behavior of c.ClientIP()
	res, err := client.Do(req)
	return res, err
}

func Test_one_client(t *testing.T) {
	// t.Skip()
	clientIP := "1.2.3.4"
	gin.SetMode(gin.ReleaseMode)

	ts := httptest.NewServer(setupServer(RequestLimitPerMinute))
	defer ts.Close()

	// consume all tokens but left one
	for i := 0; i < 58; i++ {
		sendReq(ts.URL, clientIP)
	}

	// use 59-th, OK
	resp, err := sendReq(ts.URL, clientIP)
	AssertErr(t, err)
	AssertStatus200(t, resp)
	AssertRemaining(t, resp.Header.Get("X-ratelimit-limit-remaining"), "1")
	AssertUsed(t, resp.Header.Get("X-ratelimit-limit-used"), "59")

	// use 60-th, OK
	resp, err = sendReq(ts.URL, clientIP)
	AssertErr(t, err)
	AssertStatus200(t, resp)
	AssertRemaining(t, resp.Header.Get("X-ratelimit-limit-remaining"), "0")
	AssertUsed(t, resp.Header.Get("X-ratelimit-limit-used"), "60")

	// use 61-th, NG
	resp, err = sendReq(ts.URL, clientIP)
	AssertErr(t, err)
	AssertStatusError(t, resp)
	AssertRemaining(t, resp.Header.Get("X-ratelimit-limit-remaining"), "0")
	AssertUsed(t, resp.Header.Get("X-ratelimit-limit-used"), "60")
}

func Test_Basic(t *testing.T) {
	// =================================================================
	// Referencing from https://kpat.io/2019/06/testing-with-gin/
	// =================================================================
	// t.Skip()
	gin.SetMode(gin.ReleaseMode)

	// The setupServer method, that we previously refactored
	// is injected into a test server
	ts := httptest.NewServer(setupServer(RequestLimitPerMinute))
	// Shut down the server and block until all requests have gone through
	defer ts.Close()

	// Make a request to our server with the {base url}/ping
	resp, err := sendReq(ts.URL, "1.2.3.4")
	AssertErr(t, err)
	AssertStatus200(t, resp)

	val, ok := resp.Header["Content-Type"]

	// Assert that the "content-type" header is actually set
	if !ok {
		t.Fatalf("Expected Content-Type header to be set")
	}

	// Assert that it was set as expected
	if val[0] != "application/json; charset=utf-8" {
		t.Fatalf("Expected \"application/json; charset=utf-8\", got %s", val[0])
	}

	// For debug purposes
	t.Logf("%v \n", resp.Header.Get("X-ratelimit-limit-per-minute"))
	t.Logf("%v \n", resp.Header.Get("X-ratelimit-limit-remaining"))
	t.Logf("%v \n", resp.Header.Get("X-ratelimit-limit-reset"))
	t.Logf("%v \n", resp.Header.Get("X-ratelimit-limit-used"))
}
