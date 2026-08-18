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
	busySlots := make([]bool, 12)
	activeSlots := make([]bool, 12)
	checkinSlots := make([]bool, 12)
	provisionalSlots := make([]bool, 12)
	lastFiveMinutes := input.Active != nil && input.Active.CheckinConfirmed && input.Active.End.After(input.Now) && input.Active.End.Sub(input.Now) <= 5*time.Minute
	for i := range colors {
		slotStart := input.Now.Add(time.Duration(i) * 5 * time.Minute)
		slotEnd := slotStart.Add(5 * time.Minute)
		known := input.Known[i]
		busy := input.Busy[i] || !known
		if input.ExactKnown {
			busy = inDisplayRanges(input.Now, slotStart, slotEnd, input.ExactBusy)
		} else if input.Active != nil {
			busy = false
		}
		var activeRange []TimeRange
		if input.Active != nil {
			activeRange = []TimeRange{{Start: input.Active.Start, End: input.Active.End}}
		}
		activeOverlap := inDisplayRanges(input.Now, slotStart, slotEnd, activeRange)
		activeSlots[i] = activeOverlap
		busySlots[i] = busy || activeOverlap
		checkinSlots[i] = activeOverlap && input.Active != nil && !input.Active.CheckinConfirmed
		provisionalSlots[i] = inDisplayRanges(input.Now, slotStart, slotEnd, input.Provisional)
	}

	seenBusyRange := false
	inBusyRange := false
	rangeColor := BusyRed
	for i := range colors {
		brightness := 100
		hex := FreeGreen
		if busySlots[i] {
			if activeSlots[i] {
				hex = BusyRed
			} else if input.Active != nil {
				hex = UpcomingViolet
			} else if !inBusyRange {
				rangeColor = BusyRed
				if seenBusyRange {
					rangeColor = UpcomingViolet
				}
				seenBusyRange = true
				inBusyRange = true
				hex = rangeColor
			} else {
				hex = rangeColor
			}
		} else {
			inBusyRange = false
		}
		if checkinSlots[i] {
			hex = CheckinOrange
		}
		if lastFiveMinutes && i == 0 && activeSlots[i] {
			hex = discoColor(input.Now)
		}
		if provisionalSlots[i] {
			hex = ProvisionalBlue
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

func discoColor(now time.Time) string {
	palette := []string{DiscoPink, DiscoCyan, DiscoWhite, DiscoPurple}
	return palette[(now.Second()/5)%len(palette)]
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
