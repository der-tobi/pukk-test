package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestServerPollReturnsIndividualLEDCommand(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	app := NewApp(&recordingRooms{freeBusy: "000000000000"}, AppConfig{
		ResourceID: 44,
		Now:        func() time.Time { return now },
		CommitWait: time.Hour,
		Logger:     discardLogger{},
	})
	if err := app.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/?action=poll&mac=abc", nil)
	rr := httptest.NewRecorder()

	NewServer(app, discardLogger{}).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rr.Code, rr.Body.String())
	}
	var command LEDCommand
	if err := json.Unmarshal(rr.Body.Bytes(), &command); err != nil {
		t.Fatalf("response is not LEDCommand JSON: %v", err)
	}
	if command.Command != "set_leds_individual" || len(command.LEDValues.Colors) != 12 {
		t.Fatalf("unexpected command %#v", command)
	}
	assertRedPrefix(t, command, 0)
}

func TestPollPushesIndividualLEDsToDeviceLocalAPI(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	device := &recordingDeviceController{done: make(chan struct{}, 1)}
	app := NewApp(&recordingRooms{freeBusy: "000111000000"}, AppConfig{
		ResourceID:        44,
		Now:               func() time.Time { return now },
		Logger:            discardLogger{},
		DeviceController:  device,
		DevicePushTimeout: time.Second,
	})
	if err := app.RefreshAvailability(context.Background()); err != nil {
		t.Fatalf("RefreshAvailability returned error: %v", err)
	}

	command := app.Poll("abc", "192.0.2.10")
	assertRingPattern(t, command, "GGGRRRGGGGGG")

	select {
	case <-device.done:
	case <-time.After(time.Second):
		t.Fatal("device controller was not called")
	}
	if device.ip != "192.0.2.10" {
		t.Fatalf("device ip = %q, want 192.0.2.10", device.ip)
	}
	if got := commandFromLEDValues(device.values); !sameLEDCommand(got, command) {
		t.Fatalf("pushed command = %#v, want %#v", got, command)
	}
}

func TestFutureBookingFifteenMinutesOutRendersGreenRedGreen(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	app := NewApp(&recordingRooms{freeBusy: "000111000000000"}, AppConfig{
		ResourceID: 44,
		Now:        func() time.Time { return now },
		Logger:     discardLogger{},
	})
	if err := app.RefreshAvailability(context.Background()); err != nil {
		t.Fatalf("RefreshAvailability returned error: %v", err)
	}

	command := app.Poll("abc", "192.0.2.10")

	assertRingPattern(t, command, "GGGRRRGGGGGG")
}

func TestPollUsesExactUpcomingBookingTimesToAvoidMidnightFreeBusyMisalignment(t *testing.T) {
	now := time.Date(2026, 8, 13, 23, 53, 0, 0, time.UTC)
	rooms := &recordingRooms{
		freeBusy: "111110000000000",
		bookings: []Booking{{
			ID:               7,
			ResourceID:       44,
			Start:            time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC),
			End:              time.Date(2026, 8, 14, 0, 15, 0, 0, time.UTC),
			CheckinConfirmed: true,
		}},
	}
	app := NewApp(rooms, AppConfig{
		ResourceID: 44,
		Now:        func() time.Time { return now },
		Logger:     discardLogger{},
	})
	if err := app.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	command := app.Poll("abc", "192.0.2.10")

	assertRingPattern(t, command, "GRRRGGGGGGGG")
}

func TestNewAppUsesThirtySecondPoCRoomsRefreshInterval(t *testing.T) {
	app := NewApp(&recordingRooms{}, AppConfig{ResourceID: 44, Logger: discardLogger{}})

	if got := app.RoomsRefreshInterval(); got != 30*time.Second {
		t.Fatalf("RoomsRefreshInterval = %s, want 30s", got)
	}
}

func TestServerHealthCheckIsReachableWithoutPukkQuery(t *testing.T) {
	app := NewApp(&recordingRooms{}, AppConfig{
		ResourceID: 44,
		Logger:     discardLogger{},
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	NewServer(app, discardLogger{}).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rr.Code, rr.Body.String())
	}
	if got := strings.TrimSpace(rr.Body.String()); got != "ok" {
		t.Fatalf("body = %q, want ok", got)
	}
}

func TestServerDebugStatusReportsObservedPoll(t *testing.T) {
	now := time.Date(2026, 8, 14, 9, 30, 0, 0, time.UTC)
	app := NewApp(&recordingRooms{freeBusy: "000000000000"}, AppConfig{
		ResourceID: 44,
		Now:        func() time.Time { return now },
		Logger:     discardLogger{},
	})
	if err := app.RefreshAvailability(context.Background()); err != nil {
		t.Fatalf("RefreshAvailability returned error: %v", err)
	}

	pollReq := httptest.NewRequest(http.MethodGet, "/?action=poll&mac=abc", nil)
	pollReq.RemoteAddr = "192.0.2.10:4321"
	NewServer(app, discardLogger{}).ServeHTTP(httptest.NewRecorder(), pollReq)

	req := httptest.NewRequest(http.MethodGet, "/debug/status", nil)
	rr := httptest.NewRecorder()
	NewServer(app, discardLogger{}).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rr.Code, rr.Body.String())
	}
	var status DebugStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &status); err != nil {
		t.Fatalf("debug response is not JSON: %v", err)
	}
	if status.TotalRequests != 1 {
		t.Fatalf("TotalRequests = %d, want 1", status.TotalRequests)
	}
	if status.PollRequests != 1 {
		t.Fatalf("PollRequests = %d, want 1", status.PollRequests)
	}
	if status.LastRequest.RemoteIP != "192.0.2.10" || status.LastRequest.MAC != "abc" {
		t.Fatalf("LastRequest = %#v", status.LastRequest)
	}
	if len(status.LastLEDHex) != 12 {
		t.Fatalf("LastLEDHex len = %d, want 12", len(status.LastLEDHex))
	}
	if status.LastLEDHex[0] == "" {
		t.Fatal("LastLEDHex[0] was empty")
	}
	if status.LastBlueLEDs != 0 {
		t.Fatalf("LastBlueLEDs = %d, want 0 for normal ring rendering", status.LastBlueLEDs)
	}
	if status.LastRedLEDs != 0 {
		t.Fatalf("LastRedLEDs = %d, want 0 for empty resource", status.LastRedLEDs)
	}
}

func TestShortPressCreatesAndCommitsAdHocBooking(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	rooms := &recordingRooms{freeBusy: "000000000000000"}
	app := NewApp(rooms, AppConfig{
		ResourceID: 44,
		Now:        func() time.Time { return now },
		CommitWait: time.Hour,
		Logger:     discardLogger{},
	})
	if err := app.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/?action=short_press&mac=abc", strings.NewReader(""))
	rr := httptest.NewRecorder()
	NewServer(app, discardLogger{}).ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body %s", rr.Code, rr.Body.String())
	}
	command := app.Poll("abc", "192.0.2.10")
	assertHex(t, command.LEDValues.Colors[0], "#FF0000")
	assertHex(t, command.LEDValues.Colors[1], "#FF0000")
	assertHex(t, command.LEDValues.Colors[2], "#FF0000")
	status := app.DebugStatus()
	if status.LastBlueLEDs != 0 {
		t.Fatalf("LastBlueLEDs = %d, want 0 after provisional short press", status.LastBlueLEDs)
	}
	if status.LastRedLEDs != 3 {
		t.Fatalf("LastRedLEDs = %d, want 3 after one short press", status.LastRedLEDs)
	}

	if err := app.CommitPending(context.Background()); err != nil {
		t.Fatalf("CommitPending returned error: %v", err)
	}
	if rooms.created.Start != now || rooms.created.End != now.Add(15*time.Minute) {
		t.Fatalf("created range = %s..%s", rooms.created.Start, rooms.created.End)
	}
}

func TestShortPressesShowFifteenMinuteRedStepsOnEmptyResource(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	rooms := &recordingRooms{freeBusy: "000000000000000"}
	app := NewApp(rooms, AppConfig{
		ResourceID: 44,
		Now:        func() time.Time { return now },
		CommitWait: time.Hour,
		Logger:     discardLogger{},
	})
	if err := app.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	app.handleShortPress()
	assertRedPrefix(t, app.Poll("abc", "192.0.2.10"), 3)

	app.handleShortPress()
	assertRedPrefix(t, app.Poll("abc", "192.0.2.10"), 6)

	app.handleShortPress()
	assertRedPrefix(t, app.Poll("abc", "192.0.2.10"), 9)
}

func TestActiveBookingPressExtendsToVisibleHourEdgeWhenOnlyTenMinutesFree(t *testing.T) {
	now := time.Date(2026, 8, 13, 17, 10, 0, 0, time.UTC)
	current := &Booking{
		ID:               7,
		Start:            time.Date(2026, 8, 13, 17, 0, 0, 0, time.UTC),
		End:              time.Date(2026, 8, 13, 18, 0, 0, 0, time.UTC),
		CheckinConfirmed: true,
	}
	rooms := &recordingRooms{
		freeBusy:       "111111111100000",
		currentBooking: current,
	}
	app := NewApp(rooms, AppConfig{
		ResourceID: 44,
		Now:        func() time.Time { return now },
		CommitWait: time.Hour,
		Logger:     discardLogger{},
	})
	if err := app.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	app.handleShortPress()
	assertRedPrefix(t, app.Poll("abc", "192.0.2.10"), 12)

	if err := app.CommitPending(context.Background()); err != nil {
		t.Fatalf("CommitPending returned error: %v", err)
	}
	wantEnd := now.Add(time.Hour)
	if rooms.extendedEnd != wantEnd {
		t.Fatalf("extended end = %s, want %s", rooms.extendedEnd, wantEnd)
	}
}

func TestStaleButtonTimerCannotCommitAfterLaterPress(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	rooms := &recordingRooms{freeBusy: "000000000000000"}
	app := NewApp(rooms, AppConfig{
		ResourceID: 44,
		Now:        func() time.Time { return now },
		CommitWait: time.Hour,
		Logger:     discardLogger{},
	})
	if err := app.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	app.handleShortPress()
	firstGeneration := pendingGeneration(app)
	app.handleShortPress()
	secondGeneration := pendingGeneration(app)

	if firstGeneration == secondGeneration {
		t.Fatal("button generation did not advance after second press")
	}
	if err := app.CommitPendingGeneration(context.Background(), firstGeneration); err != nil {
		t.Fatalf("stale CommitPendingGeneration returned error: %v", err)
	}
	if !rooms.created.Start.IsZero() {
		t.Fatalf("stale generation created booking %+v", rooms.created)
	}
	if err := app.CommitPendingGeneration(context.Background(), secondGeneration); err != nil {
		t.Fatalf("current CommitPendingGeneration returned error: %v", err)
	}
	if rooms.created.Start != now || rooms.created.End != now.Add(30*time.Minute) {
		t.Fatalf("created range = %s..%s, want 30 minute booking", rooms.created.Start, rooms.created.End)
	}
}

func TestCommittedAdHocBookingRendersRedBeforeFreeBusyCatchesUp(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	rooms := &recordingRooms{freeBusy: "000000000000000"}
	app := NewApp(rooms, AppConfig{
		ResourceID: 44,
		Now:        func() time.Time { return now },
		CommitWait: time.Hour,
		Logger:     discardLogger{},
	})
	if err := app.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	app.handleShortPress()
	if err := app.CommitPending(context.Background()); err != nil {
		t.Fatalf("CommitPending returned error: %v", err)
	}
	command := app.Poll("abc", "192.0.2.10")

	assertHex(t, command.LEDValues.Colors[0], "#FF0000")
	assertHex(t, command.LEDValues.Colors[1], "#FF0000")
	assertHex(t, command.LEDValues.Colors[2], "#FF0000")
	assertHex(t, command.LEDValues.Colors[3], "#00FF00")
}

func TestPollDoesNotRenderExpiredPendingSelectionBlue(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	rooms := &recordingRooms{freeBusy: "000000000000000"}
	app := NewApp(rooms, AppConfig{
		ResourceID: 44,
		Now:        func() time.Time { return now },
		CommitWait: 5 * time.Second,
		Logger:     discardLogger{},
	})
	if err := app.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	app.handleShortPress()

	now = now.Add(6 * time.Second)
	command := app.Poll("abc", "192.0.2.10")

	assertHex(t, command.LEDValues.Colors[0], "#FF0000")
	assertHex(t, command.LEDValues.Colors[1], "#FF0000")
	assertHex(t, command.LEDValues.Colors[2], "#FF0000")
	status := app.DebugStatus()
	if !status.PendingUntil.IsZero() {
		t.Fatalf("PendingUntil = %s, want cleared after expired pending fallback", status.PendingUntil)
	}
}

func TestRefreshAvailabilityStoresBitStringWithoutActiveBookingLookup(t *testing.T) {
	now := time.Date(2026, 8, 14, 10, 0, 0, 0, time.UTC)
	rooms := &recordingRooms{freeBusy: "010101010101"}
	app := NewApp(rooms, AppConfig{
		ResourceID: 44,
		Now:        func() time.Time { return now },
		Logger:     discardLogger{},
	})

	if err := app.RefreshAvailability(context.Background()); err != nil {
		t.Fatalf("RefreshAvailability returned error: %v", err)
	}
	status := app.DebugStatus()

	if rooms.findCurrentCalls != 0 {
		t.Fatalf("FindCurrentBooking was called %d times, want 0", rooms.findCurrentCalls)
	}
	if status.AvailabilityBits != expandFiveMinuteFreeBusy("010101010101") {
		t.Fatalf("AvailabilityBits = %q", status.AvailabilityBits)
	}
	if status.AvailabilityIntervalMinutes != 1 {
		t.Fatalf("AvailabilityIntervalMinutes = %d, want 1", status.AvailabilityIntervalMinutes)
	}
	if status.AvailabilityStart != now || status.AvailabilityEnd != now.Add(75*time.Minute) {
		t.Fatalf("availability window = %s..%s", status.AvailabilityStart, status.AvailabilityEnd)
	}
	if status.LastAvailabilityRefreshOK.IsZero() {
		t.Fatal("LastAvailabilityRefreshOK was not set")
	}
}

func TestRefreshLoopFetchesAvailabilityImmediately(t *testing.T) {
	rooms := &recordingRooms{
		freeBusy:      "000000000000",
		freeBusyCalls: make(chan struct{}, 1),
	}
	app := NewApp(rooms, AppConfig{
		ResourceID: 44,
		Logger:     discardLogger{},
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go refreshLoop(ctx, app, discardLogger{})

	select {
	case <-rooms.freeBusyCalls:
	case <-time.After(time.Second):
		t.Fatal("refreshLoop did not fetch availability immediately")
	}
}

func pendingGeneration(app *App) int {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.pendingGeneration
}

type recordingRooms struct {
	freeBusy         string
	freeBusyExact    bool
	freeBusyStart    time.Time
	freeBusyEnd      time.Time
	freeBusyInterval int
	created          TimeRange
	extendedID       int
	checkedInID      int
	releasedID       int
	findCurrentCalls int
	freeBusyCalls    chan struct{}
	currentBooking   *Booking
	bookings         []Booking
	extendedEnd      time.Time
}

func (r *recordingRooms) FreeBusy(_ context.Context, start, end time.Time, intervalMinutes int) (string, error) {
	if r.freeBusyCalls != nil {
		select {
		case r.freeBusyCalls <- struct{}{}:
		default:
		}
	}
	r.freeBusyStart = start
	r.freeBusyEnd = end
	r.freeBusyInterval = intervalMinutes
	if intervalMinutes == 1 && !r.freeBusyExact {
		return expandFiveMinuteFreeBusy(r.freeBusy), nil
	}
	return r.freeBusy, nil
}

func (r *recordingRooms) FindCurrentBooking(context.Context, time.Time) (*Booking, error) {
	r.findCurrentCalls++
	if r.currentBooking == nil {
		return nil, nil
	}
	booking := *r.currentBooking
	return &booking, nil
}

func (r *recordingRooms) FindBookings(_ context.Context, start, end time.Time) ([]Booking, error) {
	var bookings []Booking
	for _, booking := range r.bookings {
		if booking.Start.Before(end) && booking.End.After(start) {
			bookings = append(bookings, booking)
		}
	}
	if r.currentBooking != nil && r.currentBooking.Start.Before(end) && r.currentBooking.End.After(start) {
		booking := *r.currentBooking
		bookings = append(bookings, booking)
	}
	return bookings, nil
}

func (r *recordingRooms) CreateAdHocBooking(_ context.Context, start, end time.Time) (*Booking, error) {
	r.created = TimeRange{Start: start, End: end}
	return &Booking{ID: 100, Start: start, End: end, CheckinConfirmed: true}, nil
}

func (r *recordingRooms) ExtendBooking(_ context.Context, booking Booking, end time.Time) (*Booking, error) {
	r.extendedID = booking.ID
	r.extendedEnd = end
	return &Booking{ID: booking.ID, Start: booking.Start, End: end, CheckinConfirmed: booking.CheckinConfirmed}, nil
}

func (r *recordingRooms) CheckInBooking(_ context.Context, booking Booking) (*Booking, error) {
	r.checkedInID = booking.ID
	booking.CheckinConfirmed = true
	return &booking, nil
}

func (r *recordingRooms) ReleaseBooking(_ context.Context, booking Booking) error {
	r.releasedID = booking.ID
	return nil
}

type discardLogger struct{}

func (discardLogger) Printf(string, ...any) {}

type recordingDeviceController struct {
	done   chan struct{}
	ip     string
	values LEDValues
}

func (r *recordingDeviceController) SetIndividualLEDs(_ context.Context, ip string, values LEDValues) error {
	r.ip = ip
	r.values = values
	if r.done != nil {
		select {
		case r.done <- struct{}{}:
		default:
		}
	}
	return nil
}

func assertRedPrefix(t *testing.T, command LEDCommand, redCount int) {
	t.Helper()
	for i := 0; i < 12; i++ {
		want := "#00FF00"
		if i < redCount {
			want = "#FF0000"
		}
		assertHex(t, command.LEDValues.Colors[i], want)
	}
}

func assertRingPattern(t *testing.T, command LEDCommand, pattern string) {
	t.Helper()
	if len(command.LEDValues.Colors) != len(pattern) {
		t.Fatalf("colors len = %d, want %d", len(command.LEDValues.Colors), len(pattern))
	}
	for i, marker := range pattern {
		want := "#00FF00"
		if marker == 'R' {
			want = "#FF0000"
		}
		if got := command.LEDValues.Colors[i].Hex(); got != want {
			t.Fatalf("slot %d color = %s, want %s for pattern %s", i, got, want, pattern)
		}
	}
}

func commandFromLEDValues(values LEDValues) LEDCommand {
	return LEDCommand{Command: "set_leds_individual", LEDValues: values}
}

func sameLEDCommand(a, b LEDCommand) bool {
	if a.Command != b.Command || len(a.LEDValues.Colors) != len(b.LEDValues.Colors) || a.LEDValues.DurationMS != b.LEDValues.DurationMS {
		return false
	}
	for i := range a.LEDValues.Colors {
		if a.LEDValues.Colors[i] != b.LEDValues.Colors[i] {
			return false
		}
	}
	return true
}

func expandFiveMinuteFreeBusy(pattern string) string {
	var expanded strings.Builder
	for _, c := range pattern {
		for i := 0; i < 5; i++ {
			expanded.WriteRune(c)
		}
	}
	return expanded.String()
}
