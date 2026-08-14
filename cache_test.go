package main

import (
	"context"
	"testing"
	"time"
)

func TestBookingCacheRefreshFetchesDisplayWindowWithRefreshAndSafetyBuffer(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 2, 30, 0, time.UTC)
	rooms := &recordingRooms{freeBusy: "000000000000000"}
	cache := NewBookingCache(rooms, CacheConfig{
		ResourceID:      44,
		DisplayWindow:   time.Hour,
		RefreshInterval: 5 * time.Minute,
		SafetyMargin:    10 * time.Minute,
		IntervalMinutes: 5,
	})

	if err := cache.Refresh(context.Background(), now); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	if !rooms.freeBusyStart.Equal(time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("start = %s", rooms.freeBusyStart)
	}
	if !rooms.freeBusyEnd.Equal(time.Date(2026, 8, 13, 11, 15, 0, 0, time.UTC)) {
		t.Fatalf("end = %s", rooms.freeBusyEnd)
	}
	if rooms.freeBusyInterval != 5 {
		t.Fatalf("interval = %d, want 5", rooms.freeBusyInterval)
	}
}

func TestBookingCacheSnapshotReturnsBusyForUnknownSlots(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	cache := NewBookingCache(&recordingRooms{freeBusy: "010"}, CacheConfig{
		ResourceID:      44,
		DisplayWindow:   time.Hour,
		RefreshInterval: 5 * time.Minute,
		SafetyMargin:    10 * time.Minute,
		IntervalMinutes: 5,
	})
	if err := cache.Refresh(context.Background(), now); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	snapshot := cache.Snapshot(now)

	if snapshot.Busy[0] {
		t.Fatal("slot 0 should be free")
	}
	if !snapshot.Busy[1] {
		t.Fatal("slot 1 should be busy")
	}
	if !snapshot.Busy[3] {
		t.Fatal("slot 3 should default to busy because cache has no data")
	}
	if snapshot.Known[3] {
		t.Fatal("slot 3 should be unknown")
	}
}

func TestBookingCacheSnapshotIgnoresElapsedCurrentSlotMinutesWithOneMinuteSource(t *testing.T) {
	refreshTime := time.Date(2026, 8, 13, 23, 0, 0, 0, time.UTC)
	now := time.Date(2026, 8, 13, 23, 18, 0, 0, time.UTC)
	rooms := &recordingRooms{
		freeBusy:      minuteFreeBusy(75, 15, 16, 17, 25),
		freeBusyExact: true,
	}
	cache := NewBookingCache(rooms, CacheConfig{
		ResourceID:      44,
		DisplayWindow:   time.Hour,
		RefreshInterval: 5 * time.Minute,
		SafetyMargin:    10 * time.Minute,
		IntervalMinutes: 1,
	})
	if err := cache.Refresh(context.Background(), refreshTime); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	snapshot := cache.Snapshot(now)

	if snapshot.Busy[0] {
		t.Fatal("slot 0 should be free because only elapsed minutes in the 23:15-23:20 bucket were busy")
	}
	if !snapshot.Busy[1] {
		t.Fatal("slot 1 should be busy because its midpoint is busy")
	}
}

func TestBookingCacheSnapshotSamplesSlotMidpointsForFutureFreeBusy(t *testing.T) {
	refreshTime := time.Date(2026, 8, 13, 23, 50, 0, 0, time.UTC)
	now := time.Date(2026, 8, 13, 23, 53, 0, 0, time.UTC)
	rooms := &recordingRooms{
		freeBusy:      minuteFreeBusy(75, rangeInts(10, 24)...),
		freeBusyExact: true,
	}
	cache := NewBookingCache(rooms, CacheConfig{
		ResourceID:      44,
		DisplayWindow:   time.Hour,
		RefreshInterval: 5 * time.Minute,
		SafetyMargin:    10 * time.Minute,
		IntervalMinutes: 1,
	})
	if err := cache.Refresh(context.Background(), refreshTime); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	snapshot := cache.Snapshot(now)

	assertSnapshotPattern(t, snapshot, "GRRRGGGGGGGG")
}

func TestNewAppRequestsOneMinuteRoomsFreeBusyResolution(t *testing.T) {
	now := time.Date(2026, 8, 13, 23, 0, 0, 0, time.UTC)
	rooms := &recordingRooms{freeBusy: minuteFreeBusy(75), freeBusyExact: true}
	app := NewApp(rooms, AppConfig{
		ResourceID: 44,
		Now:        func() time.Time { return now },
		Logger:     discardLogger{},
	})

	if err := app.RefreshAvailability(context.Background()); err != nil {
		t.Fatalf("RefreshAvailability returned error: %v", err)
	}

	if rooms.freeBusyInterval != 1 {
		t.Fatalf("rooms freebusy interval = %d, want 1", rooms.freeBusyInterval)
	}
}

func minuteFreeBusy(length int, busyIndexes ...int) string {
	data := make([]byte, length)
	for i := range data {
		data[i] = '0'
	}
	for _, idx := range busyIndexes {
		if idx >= 0 && idx < len(data) {
			data[idx] = '1'
		}
	}
	return string(data)
}

func rangeInts(start, endInclusive int) []int {
	var indexes []int
	for i := start; i <= endInclusive; i++ {
		indexes = append(indexes, i)
	}
	return indexes
}

func assertSnapshotPattern(t *testing.T, snapshot AvailabilitySnapshot, pattern string) {
	t.Helper()
	for i, marker := range pattern {
		wantBusy := marker == 'R'
		if snapshot.Busy[i] != wantBusy {
			t.Fatalf("slot %d busy = %v, want %v for pattern %s", i, snapshot.Busy[i], wantBusy, pattern)
		}
	}
}
