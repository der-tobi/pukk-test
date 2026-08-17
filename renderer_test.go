package main

import (
	"testing"
	"time"
)

func TestRenderRingColorsFirstBusyRangeRedAndLaterBusyRangesViolet(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 2, 0, 0, time.UTC)
	input := RingInput{
		Now:        now,
		Known:      slots("111111111111"),
		ExactKnown: true,
		ExactBusy: []TimeRange{
			{Start: now, End: now.Add(15 * time.Minute)},
			{Start: now.Add(30 * time.Minute), End: now.Add(45 * time.Minute)},
		},
	}

	command := RenderRing(input)

	if command.Command != "set_leds_individual" {
		t.Fatalf("command = %q, want set_leds_individual", command.Command)
	}
	if len(command.LEDValues.Colors) != 12 {
		t.Fatalf("colors len = %d, want 12", len(command.LEDValues.Colors))
	}
	assertHex(t, command.LEDValues.Colors[0], "#FF0000")
	assertHex(t, command.LEDValues.Colors[1], "#FF0000")
	assertHex(t, command.LEDValues.Colors[2], "#FF0000")
	assertHex(t, command.LEDValues.Colors[3], "#00FF00")
	assertHex(t, command.LEDValues.Colors[6], UpcomingViolet)
	assertHex(t, command.LEDValues.Colors[7], UpcomingViolet)
	assertHex(t, command.LEDValues.Colors[8], UpcomingViolet)
}

func TestRenderRingTreatsUnknownSlotsAsBusy(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	command := RenderRing(RingInput{Now: now})

	assertHex(t, command.LEDValues.Colors[0], "#FF0000")
	assertHex(t, command.LEDValues.Colors[11], "#FF0000")
}

func TestRenderRingOverlaysProvisionalBlocksInBlue(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	command := RenderRing(RingInput{
		Now:   now,
		Busy:  slots("000000000000"),
		Known: slots("111111111111"),
		Provisional: []TimeRange{{
			Start: now,
			End:   now.Add(15 * time.Minute),
		}},
	})

	assertHex(t, command.LEDValues.Colors[0], ProvisionalBlue)
	assertHex(t, command.LEDValues.Colors[1], ProvisionalBlue)
	assertHex(t, command.LEDValues.Colors[2], ProvisionalBlue)
	assertHex(t, command.LEDValues.Colors[3], "#00FF00")
}

func TestRenderRingShowsCheckedInCurrentBookingAsRed(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	command := RenderRing(RingInput{
		Now:   now,
		Busy:  slots("111000000000"),
		Known: slots("111111111111"),
		Active: &Booking{
			ID:               7,
			Start:            now,
			End:              now.Add(15 * time.Minute),
			CheckinConfirmed: true,
		},
	})

	assertHex(t, command.LEDValues.Colors[0], BusyRed)
	assertHex(t, command.LEDValues.Colors[1], BusyRed)
	assertHex(t, command.LEDValues.Colors[2], BusyRed)
}

func TestRenderRingShowsOnlyUncheckedCurrentBookingLEDsAsOrange(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	command := RenderRing(RingInput{
		Now:        now,
		Known:      slots("111111111111"),
		ExactKnown: true,
		ExactBusy: []TimeRange{
			{Start: now, End: now.Add(15 * time.Minute)},
			{Start: now.Add(30 * time.Minute), End: now.Add(45 * time.Minute)},
		},
		Active: &Booking{
			ID:               7,
			Start:            now,
			End:              now.Add(15 * time.Minute),
			CheckinConfirmed: false,
		},
	})

	assertHex(t, command.LEDValues.Colors[0], CheckinOrange)
	assertHex(t, command.LEDValues.Colors[1], CheckinOrange)
	assertHex(t, command.LEDValues.Colors[2], CheckinOrange)
	assertHex(t, command.LEDValues.Colors[3], FreeGreen)
	assertHex(t, command.LEDValues.Colors[6], UpcomingViolet)
	assertHex(t, command.LEDValues.Colors[7], UpcomingViolet)
	assertHex(t, command.LEDValues.Colors[8], UpcomingViolet)
}

func TestRenderRingDoesNotLetClosedHoursFallbackExpandCurrentCheckinBooking(t *testing.T) {
	now := time.Date(2026, 8, 13, 23, 42, 0, 0, time.UTC)
	command := RenderRing(RingInput{
		Now:   now,
		Busy:  slots("111111111111"),
		Known: slots("111111111111"),
		Active: &Booking{
			ID:               7,
			Start:            time.Date(2026, 8, 13, 23, 30, 0, 0, time.UTC),
			End:              time.Date(2026, 8, 13, 23, 45, 0, 0, time.UTC),
			CheckinConfirmed: false,
		},
	})

	assertHex(t, command.LEDValues.Colors[0], CheckinOrange)
	for i := 1; i < 12; i++ {
		assertHex(t, command.LEDValues.Colors[i], FreeGreen)
	}
}

func TestRenderRingDoesNotLetClosedHoursFallbackExpandCheckedInCurrentBooking(t *testing.T) {
	now := time.Date(2026, 8, 13, 23, 42, 0, 0, time.UTC)
	command := RenderRing(RingInput{
		Now:   now,
		Busy:  slots("111111111111"),
		Known: slots("111111111111"),
		Active: &Booking{
			ID:               7,
			Start:            time.Date(2026, 8, 13, 23, 30, 0, 0, time.UTC),
			End:              time.Date(2026, 8, 13, 23, 45, 0, 0, time.UTC),
			CheckinConfirmed: true,
		},
	})

	assertHex(t, command.LEDValues.Colors[0], BusyRed)
	for i := 1; i < 12; i++ {
		assertHex(t, command.LEDValues.Colors[i], FreeGreen)
	}
}

func slots(pattern string) [12]bool {
	var result [12]bool
	for i, c := range pattern {
		if i >= len(result) {
			break
		}
		result[i] = c == '1'
	}
	return result
}

func assertHex(t *testing.T, color LEDColor, want string) {
	t.Helper()
	if got := color.Hex(); got != want {
		t.Fatalf("color = %s, want %s", got, want)
	}
}
