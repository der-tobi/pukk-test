package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestRoomsClientAuthenticatesAndFetchesFreeBusy(t *testing.T) {
	var tokenForm url.Values
	var freeBusyAuth string
	var freeBusyPath string
	var freeBusyQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/connect/token":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm: %v", err)
			}
			tokenForm = r.PostForm
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "token-1"})
		case "/default/api/v2.0/ressources/44/freebusy":
			freeBusyAuth = r.Header.Get("Authorization")
			freeBusyPath = r.URL.Path
			freeBusyQuery = r.URL.Query()
			_ = json.NewEncoder(w).Encode(map[string]any{"FreeBusyData": "010101", "ResourceId": 44})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewRoomsHTTPClient(RoomsConfig{
		AuthBaseURL: server.URL,
		APIBaseURL:  server.URL,
		Mandator:    "default",
		ResourceID:  44,
		User:        "tobiapi4",
		Password:    "secret",
		HTTPClient:  server.Client(),
	})

	got, err := client.FreeBusy(context.Background(), time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC), time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC), 5)
	if err != nil {
		t.Fatalf("FreeBusy returned error: %v", err)
	}

	if got != "010101" {
		t.Fatalf("FreeBusyData = %q", got)
	}
	if tokenForm.Get("grant_type") != "basic" || tokenForm.Get("client_id") != "basic-auth" || tokenForm.Get("scope") != "rooms_api" || tokenForm.Get("user") != "tobiapi4" || tokenForm.Get("password") != "secret" {
		t.Fatalf("unexpected token form %v", tokenForm)
	}
	if freeBusyAuth != "Bearer token-1" {
		t.Fatalf("Authorization = %q", freeBusyAuth)
	}
	if freeBusyPath != "/default/api/v2.0/ressources/44/freebusy" {
		t.Fatalf("path = %q", freeBusyPath)
	}
	if freeBusyQuery.Get("interval") != "5" || !strings.Contains(freeBusyQuery.Get("start"), "2026-08-13T10:00:00") {
		t.Fatalf("query = %v", freeBusyQuery)
	}
}

func TestRoomsClientFindBookingsUsesResourceAndWindowFilters(t *testing.T) {
	var bookingAuth string
	var bookingQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/connect/token":
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "token-1"})
		case "/default/api/v2.0/bookings/find":
			bookingAuth = r.Header.Get("Authorization")
			bookingQuery = r.URL.Query()
			_ = json.NewEncoder(w).Encode([]Booking{{
				ID:         7,
				ResourceID: 44,
				Start:      time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC),
				End:        time.Date(2026, 8, 14, 0, 15, 0, 0, time.UTC),
			}})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewRoomsHTTPClient(RoomsConfig{
		AuthBaseURL: server.URL,
		APIBaseURL:  server.URL,
		Mandator:    "default",
		ResourceID:  44,
		User:        "tobiapi4",
		Password:    "secret",
		HTTPClient:  server.Client(),
	})
	start := time.Date(2026, 8, 13, 23, 45, 0, 0, time.UTC)
	end := time.Date(2026, 8, 14, 0, 45, 0, 0, time.UTC)

	bookings, err := client.FindBookings(context.Background(), start, end)
	if err != nil {
		t.Fatalf("FindBookings returned error: %v", err)
	}

	if bookingAuth != "Bearer token-1" {
		t.Fatalf("Authorization = %q", bookingAuth)
	}
	if bookingQuery.Get("filter.resourceId") != "44" || !strings.Contains(bookingQuery.Get("filter.start"), "2026-08-13T23:45:00") || !strings.Contains(bookingQuery.Get("filter.end"), "2026-08-14T00:45:00") {
		t.Fatalf("query = %v", bookingQuery)
	}
	if len(bookings) != 1 || bookings[0].BookingID() != 7 {
		t.Fatalf("bookings = %#v", bookings)
	}
}
