package main

import (
	"context"
	"sync"
	"time"
)

type FreeBusyReader interface {
	FreeBusy(ctx context.Context, start, end time.Time, intervalMinutes int) (string, error)
}

type CacheConfig struct {
	ResourceID      int
	DisplayWindow   time.Duration
	RefreshInterval time.Duration
	SafetyMargin    time.Duration
	IntervalMinutes int
}

type BookingCache struct {
	rooms  FreeBusyReader
	config CacheConfig

	mu     sync.RWMutex
	window cachedFreeBusy
}

type cachedFreeBusy struct {
	start           time.Time
	end             time.Time
	intervalMinutes int
	data            string
}

type FreeBusyWindow struct {
	Start           time.Time `json:"start"`
	End             time.Time `json:"end"`
	IntervalMinutes int       `json:"intervalMinutes"`
	Data            string    `json:"data"`
}

type AvailabilitySnapshot struct {
	Now   time.Time
	Busy  [12]bool
	Known [12]bool
}

func NewBookingCache(rooms FreeBusyReader, config CacheConfig) *BookingCache {
	if config.DisplayWindow == 0 {
		config.DisplayWindow = time.Hour
	}
	if config.RefreshInterval == 0 {
		config.RefreshInterval = 5 * time.Minute
	}
	if config.SafetyMargin == 0 {
		config.SafetyMargin = 10 * time.Minute
	}
	if config.IntervalMinutes == 0 {
		config.IntervalMinutes = 5
	}
	return &BookingCache{rooms: rooms, config: config}
}

func (c *BookingCache) Refresh(ctx context.Context, now time.Time) error {
	start := bucketStart(now)
	end := start.Add(c.config.DisplayWindow + c.config.RefreshInterval + c.config.SafetyMargin)
	data, err := c.rooms.FreeBusy(ctx, start, end, c.config.IntervalMinutes)
	if err != nil {
		return err
	}
	intervalMinutes := inferSourceIntervalMinutes(start, end, data, c.config.IntervalMinutes)

	c.mu.Lock()
	c.window = cachedFreeBusy{
		start:           start,
		end:             end,
		intervalMinutes: intervalMinutes,
		data:            data,
	}
	c.mu.Unlock()
	return nil
}

func (c *BookingCache) Window() FreeBusyWindow {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return FreeBusyWindow{
		Start:           c.window.start,
		End:             c.window.end,
		IntervalMinutes: c.window.intervalMinutes,
		Data:            c.window.data,
	}
}

func (c *BookingCache) Snapshot(now time.Time) AvailabilitySnapshot {
	c.mu.RLock()
	window := c.window
	c.mu.RUnlock()

	var snapshot AvailabilitySnapshot
	snapshot.Now = now
	sourceInterval := time.Duration(window.intervalMinutes) * time.Minute
	if sourceInterval <= 0 {
		sourceInterval = 5 * time.Minute
	}
	displayInterval := 5 * time.Minute

	for i := 0; i < 12; i++ {
		slotStart := now.Add(time.Duration(i) * displayInterval)
		slotEnd := slotStart.Add(displayInterval)
		slotMidpoint := slotStart.Add(slotEnd.Sub(slotStart) / 2)
		snapshot.Known[i], snapshot.Busy[i] = window.DisplayState(slotMidpoint, displayInterval, sourceInterval)
	}
	return snapshot
}

func (w cachedFreeBusy) DisplayState(at time.Time, displayInterval, sourceInterval time.Duration) (bool, bool) {
	known, busy := w.PointState(at, sourceInterval)
	if !known || !busy || sourceInterval <= displayInterval {
		return known, busy
	}
	idx := int(at.Sub(w.start) / sourceInterval)
	bucketStart := w.start.Add(time.Duration(idx) * sourceInterval)
	bucketEnd := bucketStart.Add(sourceInterval)
	previousBusy := idx > 0 && w.data[idx-1] == '1'
	nextBusy := idx+1 < len(w.data) && w.data[idx+1] == '1'

	if !previousBusy && at.Sub(bucketStart) < sourceInterval-displayInterval {
		return true, false
	}
	if !nextBusy && bucketEnd.Sub(at) <= displayInterval {
		return true, false
	}
	return true, true
}

func (w cachedFreeBusy) PointState(at time.Time, interval time.Duration) (bool, bool) {
	if interval <= 0 {
		return false, true
	}
	if at.Before(w.start) || !at.Before(w.end) {
		return false, true
	}
	idx := int(at.Sub(w.start) / interval)
	if idx < 0 || idx >= len(w.data) {
		return false, true
	}
	return true, w.data[idx] == '1'
}

func inferSourceIntervalMinutes(start, end time.Time, data string, requested int) int {
	if requested <= 0 {
		requested = 5
	}
	if len(data) == 0 || !end.After(start) {
		return requested
	}
	windowMinutes := int(end.Sub(start) / time.Minute)
	if windowMinutes <= 0 || windowMinutes%len(data) != 0 {
		return requested
	}
	inferred := windowMinutes / len(data)
	if inferred <= 0 {
		return requested
	}
	if !isPlausibleRoomsInterval(inferred) {
		return requested
	}
	return inferred
}

func isPlausibleRoomsInterval(minutes int) bool {
	switch minutes {
	case 1, 5, 15:
		return true
	default:
		return false
	}
}
