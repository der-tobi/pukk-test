package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultAuthBaseURL = "https://idp.vnext.3vrooms.app"
	DefaultAPIBaseURL  = "https://book.vnext.3vrooms.app"
	DefaultMandator    = "default"
	DefaultResourceID  = 44
	DefaultRoomsUser   = "tobiapi4"
)

type RoomsConfig struct {
	AuthBaseURL string
	APIBaseURL  string
	Mandator    string
	ResourceID  int
	User        string
	Password    string
	HTTPClient  *http.Client
}

type RoomsHTTPClient struct {
	cfg        RoomsConfig
	httpClient *http.Client

	mu    sync.Mutex
	token string
}

func NewRoomsHTTPClient(cfg RoomsConfig) *RoomsHTTPClient {
	if cfg.AuthBaseURL == "" {
		cfg.AuthBaseURL = DefaultAuthBaseURL
	}
	if cfg.APIBaseURL == "" {
		cfg.APIBaseURL = DefaultAPIBaseURL
	}
	if cfg.Mandator == "" {
		cfg.Mandator = DefaultMandator
	}
	if cfg.ResourceID == 0 {
		cfg.ResourceID = DefaultResourceID
	}
	if cfg.User == "" {
		cfg.User = DefaultRoomsUser
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &RoomsHTTPClient{cfg: cfg, httpClient: httpClient}
}

func (c *RoomsHTTPClient) FreeBusy(ctx context.Context, start, end time.Time, intervalMinutes int) (string, error) {
	values := url.Values{}
	values.Set("start", formatRoomsTime(start))
	values.Set("end", formatRoomsTime(end))
	values.Set("interval", strconv.Itoa(intervalMinutes))

	var dto struct {
		FreeBusyData string `json:"FreeBusyData"`
		ResourceID   int    `json:"ResourceId"`
	}
	if err := c.doJSON(ctx, http.MethodGet, c.apiPath(fmt.Sprintf("/api/v2.0/ressources/%d/freebusy", c.cfg.ResourceID), values), nil, &dto); err != nil {
		return "", err
	}
	return dto.FreeBusyData, nil
}

func (c *RoomsHTTPClient) FindCurrentBooking(ctx context.Context, now time.Time) (*Booking, error) {
	bookings, err := c.FindBookings(ctx, now.Add(-2*time.Minute), now.Add(2*time.Minute))
	if err != nil {
		return nil, err
	}
	for _, booking := range bookings {
		if booking.MatchesResource(c.cfg.ResourceID) && !now.Before(booking.Start) && now.Before(booking.End) {
			b := booking
			return &b, nil
		}
	}
	return nil, nil
}

func (c *RoomsHTTPClient) FindBookings(ctx context.Context, start, end time.Time) ([]Booking, error) {
	values := url.Values{}
	values.Set("filter.resourceId", strconv.Itoa(c.cfg.ResourceID))
	values.Set("filter.start", formatRoomsTime(start))
	values.Set("filter.end", formatRoomsTime(end))

	var bookings []Booking
	if err := c.doJSON(ctx, http.MethodGet, c.apiPath("/api/v2.0/bookings/find", values), nil, &bookings); err != nil {
		var paged struct {
			Items []Booking `json:"items"`
		}
		if err2 := c.doJSON(ctx, http.MethodGet, c.apiPath("/api/v2.0/bookings/find", values), nil, &paged); err2 != nil {
			return nil, err
		}
		bookings = paged.Items
	}
	return bookings, nil
}

func (c *RoomsHTTPClient) CreateAdHocBooking(ctx context.Context, start, end time.Time) (*Booking, error) {
	body := map[string]any{
		"ressourceId":     c.cfg.ResourceID,
		"start":           formatRoomsTime(start),
		"end":             formatRoomsTime(end),
		"title":           "PuKK Ad-hoc",
		"comment":         "Created by PuKK PoC",
		"headCount":       1,
		"isPrivate":       false,
		"equipmentList":   []any{},
		"participantList": []any{},
	}
	var booking Booking
	if err := c.doJSON(ctx, http.MethodPost, c.apiPath("/api/v2.0/bookings/quickbooking", nil), body, &booking); err != nil {
		return nil, err
	}
	return &booking, nil
}

func (c *RoomsHTTPClient) ExtendBooking(ctx context.Context, booking Booking, end time.Time) (*Booking, error) {
	id := booking.BookingID()
	if id == 0 {
		return nil, errors.New("cannot extend booking without id")
	}
	values := url.Values{}
	values.Set("end", formatRoomsTime(end))
	var updated Booking
	if err := c.doJSON(ctx, http.MethodPut, c.apiPath(fmt.Sprintf("/api/v1.0/bookings/%d", id), values), nil, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (c *RoomsHTTPClient) CheckInBooking(ctx context.Context, booking Booking) (*Booking, error) {
	id := booking.BookingID()
	if id == 0 {
		return nil, errors.New("cannot check in booking without id")
	}
	var checkedIn Booking
	if err := c.doJSON(ctx, http.MethodPut, c.apiPath(fmt.Sprintf("/api/v1.0/bookings/%d/checkin", id), nil), nil, &checkedIn); err != nil {
		return nil, err
	}
	return &checkedIn, nil
}

func (c *RoomsHTTPClient) ReleaseBooking(ctx context.Context, booking Booking) error {
	id := booking.BookingID()
	if id == 0 {
		return errors.New("cannot release booking without id")
	}
	canCheckout, err := c.canCheckout(ctx, id)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/api/v1.0/bookings/%d/release", id)
	values := url.Values{}
	if canCheckout {
		path = fmt.Sprintf("/api/v1.0/bookings/%d/checkout", id)
		values.Set("sendMail", "false")
	}
	var ignored Booking
	return c.doJSON(ctx, http.MethodPut, c.apiPath(path, values), nil, &ignored)
}

func (c *RoomsHTTPClient) canCheckout(ctx context.Context, id int) (bool, error) {
	var result bool
	if err := c.doJSON(ctx, http.MethodGet, c.apiPath(fmt.Sprintf("/api/v1.0/bookings/%d/cancheckout", id), nil), nil, &result); err != nil {
		return false, err
	}
	return result, nil
}

func (c *RoomsHTTPClient) doJSON(ctx context.Context, method, rawURL string, body any, out any) error {
	if err := c.doJSONOnce(ctx, method, rawURL, body, out); err != errUnauthorized {
		return err
	}
	c.mu.Lock()
	c.token = ""
	c.mu.Unlock()
	return c.doJSONOnce(ctx, method, rawURL, body, out)
}

var errUnauthorized = errors.New("rooms unauthorized")

func (c *RoomsHTTPClient) doJSONOnce(ctx context.Context, method, rawURL string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	token, err := c.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return errUnauthorized
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("rooms %s %s returned %s: %s", method, rawURL, resp.Status, strings.TrimSpace(string(body)))
	}
	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *RoomsHTTPClient) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.token != "" {
		token := c.token
		c.mu.Unlock()
		return token, nil
	}
	c.mu.Unlock()

	form := url.Values{}
	form.Set("grant_type", "basic")
	form.Set("client_id", "basic-auth")
	form.Set("scope", "rooms_api")
	form.Set("user", c.cfg.User)
	form.Set("password", c.cfg.Password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.cfg.AuthBaseURL, "/")+"/connect/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("rooms token returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var dto struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return "", err
	}
	if dto.AccessToken == "" {
		return "", errors.New("rooms token response did not contain access_token")
	}

	c.mu.Lock()
	c.token = dto.AccessToken
	c.mu.Unlock()
	return dto.AccessToken, nil
}

func (c *RoomsHTTPClient) apiPath(path string, values url.Values) string {
	base := strings.TrimRight(c.cfg.APIBaseURL, "/")
	full := fmt.Sprintf("%s/%s%s", base, strings.Trim(c.cfg.Mandator, "/"), path)
	if len(values) > 0 {
		full += "?" + values.Encode()
	}
	return full
}

func formatRoomsTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
