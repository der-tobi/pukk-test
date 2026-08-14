package main

import "time"

type RingInput struct {
	Now         time.Time
	Busy        [12]bool
	Known       [12]bool
	Active      *Booking
	ExactBusy   []TimeRange
	ExactKnown  bool
	Provisional []TimeRange
}

func RenderRing(input RingInput) LEDCommand {
	colors := make([]LEDColor, 12)
	for i := range colors {
		brightness := 100
		hex := FreeGreen
		slotStart := input.Now.Add(time.Duration(i) * 5 * time.Minute)
		slotEnd := slotStart.Add(5 * time.Minute)
		known := input.Known[i]
		busy := input.Busy[i] || !known
		if input.ExactKnown {
			busy = inDisplayRanges(input.Now, slotStart, slotEnd, input.ExactBusy)
		}
		var activeRange []TimeRange
		if input.Active != nil {
			activeRange = []TimeRange{{Start: input.Active.Start, End: input.Active.End}}
		}
		activeOverlap := inDisplayRanges(input.Now, slotStart, slotEnd, activeRange)
		provisionalOverlap := inDisplayRanges(input.Now, slotStart, slotEnd, input.Provisional)

		if busy || activeOverlap || provisionalOverlap {
			hex = BusyRed
		}

		colors[i] = ColorFromHex(hex, brightness)
	}

	return LEDCommand{
		Command: "set_leds_individual",
		LEDValues: LEDValues{
			Colors:     colors,
			DurationMS: 0,
		},
	}
}

func inDisplayRanges(now, slotStart, slotEnd time.Time, ranges []TimeRange) bool {
	slotMidpoint := slotStart.Add(slotEnd.Sub(slotStart) / 2)
	for _, r := range ranges {
		if !r.Start.After(now) {
			if slotStart.Before(r.End) && slotEnd.After(r.Start) {
				return true
			}
			continue
		}
		if !slotMidpoint.Before(r.Start) && slotMidpoint.Before(r.End) {
			return true
		}
	}
	return false
}

func bucketStart(t time.Time) time.Time {
	return t.Truncate(5 * time.Minute)
}
