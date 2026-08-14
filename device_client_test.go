package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestPukkDeviceHTTPClientPostsIndividualLEDValues(t *testing.T) {
	var capturedRequest *http.Request
	var capturedBody string
	client := &PukkDeviceHTTPClient{httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		capturedRequest = req
		body, _ := io.ReadAll(req.Body)
		capturedBody = string(body)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"status":"OK"}`)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})}}
	values := LEDValues{
		Colors:     []LEDColor{ColorFromHex(BusyRed, 100), ColorFromHex(FreeGreen, 100)},
		DurationMS: 0,
	}

	if err := client.SetIndividualLEDs(context.Background(), "192.0.2.10", values); err != nil {
		t.Fatalf("SetIndividualLEDs returned error: %v", err)
	}

	if capturedRequest.Method != http.MethodPost {
		t.Fatalf("method = %s, want POST", capturedRequest.Method)
	}
	if got := capturedRequest.URL.String(); got != "http://192.0.2.10/api/setLeds/individual" {
		t.Fatalf("url = %s", got)
	}
	var payload struct {
		Command   string    `json:"command"`
		LEDValues LEDValues `json:"led_values"`
	}
	if err := json.Unmarshal([]byte(capturedBody), &payload); err != nil {
		t.Fatalf("payload is not JSON: %v", err)
	}
	if payload.Command != "" {
		t.Fatalf("local REST payload command = %q, want empty", payload.Command)
	}
	if len(payload.LEDValues.Colors) != 2 {
		t.Fatalf("colors len = %d, want 2", len(payload.LEDValues.Colors))
	}
	assertHex(t, payload.LEDValues.Colors[0], BusyRed)
	assertHex(t, payload.LEDValues.Colors[1], FreeGreen)
}

func TestPukkDeviceHTTPClientSkipsLoopbackAddress(t *testing.T) {
	called := false
	client := &PukkDeviceHTTPClient{httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		called = true
		return nil, nil
	})}}

	if err := client.SetIndividualLEDs(context.Background(), "127.0.0.1", LEDValues{}); err != nil {
		t.Fatalf("SetIndividualLEDs returned error: %v", err)
	}
	if called {
		t.Fatal("loopback address should not be pushed")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
