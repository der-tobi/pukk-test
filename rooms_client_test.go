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
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"Id":          7,
				"RessourceId": 44,
				"Begin":       "2026-08-14T00:00:00Z",
				"End":         "2026-08-14T00:15:00Z",
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
	if got := bookings[0].Start; !got.Equal(time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("booking start = %s", got)
	}
}

func TestRoomsClientFindBookingsDecodesPagedPascalGermanFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/connect/token":
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "token-1"})
		case "/default/api/v2.0/bookings/find":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"Items": []map[string]any{{
					"Id":               7,
					"RessourceId":      44,
					"Beginn":           "2026-08-14T01:15:00Z",
					"Ende":             "2026-08-14T01:45:00Z",
					"CheckinConfirmed": true,
				}},
			})
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

	bookings, err := client.FindBookings(
		context.Background(),
		time.Date(2026, 8, 14, 1, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 14, 2, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("FindBookings returned error: %v", err)
	}

	if len(bookings) != 1 {
		t.Fatalf("bookings len = %d, want 1: %#v", len(bookings), bookings)
	}
	if !bookings[0].MatchesResource(44) {
		t.Fatalf("booking resource ids = resourceId %d, ressourceId %d", bookings[0].ResourceID, bookings[0].RessourceID)
	}
	if got := bookings[0].Start; !got.Equal(time.Date(2026, 8, 14, 1, 15, 0, 0, time.UTC)) {
		t.Fatalf("booking start = %s", got)
	}
	if got := bookings[0].End; !got.Equal(time.Date(2026, 8, 14, 1, 45, 0, 0, time.UTC)) {
		t.Fatalf("booking end = %s", got)
	}
}

func TestRoomsClientCheckInBookingUsesAuthenticatedUserWithoutPin(t *testing.T) {
	var checkinMethod string
	var checkinPath string
	var checkinQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/connect/token":
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "token-1"})
		case "/default/api/v1.0/bookings/7/checkin":
			checkinMethod = r.Method
			checkinPath = r.URL.Path
			checkinQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode(map[string]any{
				"Id":               7,
				"RessourceId":      44,
				"Begin":            "2026-08-13T10:00:00Z",
				"End":              "2026-08-13T10:15:00Z",
				"CheckinConfirmed": true,
			})
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

	booking, err := client.CheckInBooking(context.Background(), Booking{ID: 7})
	if err != nil {
		t.Fatalf("CheckInBooking returned error: %v", err)
	}

	if checkinMethod != http.MethodPut {
		t.Fatalf("method = %q, want PUT", checkinMethod)
	}
	if checkinPath != "/default/api/v1.0/bookings/7/checkin" {
		t.Fatalf("path = %q", checkinPath)
	}
	if checkinQuery != "" {
		t.Fatalf("query = %q, want no pin query parameter", checkinQuery)
	}
	if booking == nil || !booking.CheckinConfirmed {
		t.Fatalf("checked-in booking = %#v", booking)
	}
}
