package main

import (
	"fmt"
	"strings"
	"time"
)

const (
	BusyRed         = "#FF0000"
	FreeGreen       = "#00FF00"
	ProvisionalBlue = "#006DFF"
)

type LEDColor struct {
	Brightness int `json:"brightness"`
	Red        int `json:"red"`
	Green      int `json:"green"`
	Blue       int `json:"blue"`
}

func ColorFromHex(hex string, brightness int) LEDColor {
	hex = strings.TrimPrefix(hex, "#")
	var red, green, blue int
	_, _ = fmt.Sscanf(hex, "%02x%02x%02x", &red, &green, &blue)
	return LEDColor{Brightness: brightness, Red: red, Green: green, Blue: blue}
}

func (c LEDColor) Hex() string {
	hex := fmt.Sprintf("#%02X%02X%02X", c.Red, c.Green, c.Blue)
	switch hex {
	case "#FF0000":
		return BusyRed
	case "#00FF00":
		return FreeGreen
	case "#006DFF":
		return ProvisionalBlue
	default:
		return hex
	}
}

type LEDValues struct {
	Colors     []LEDColor `json:"colors"`
	DurationMS int        `json:"duration_ms"`
}

type LEDCommand struct {
	Command   string    `json:"command"`
	LEDValues LEDValues `json:"led_values"`
}

type Booking struct {
	ID               int       `json:"id"`
	ReservationID    int       `json:"reservationId,omitempty"`
	ResourceID       int       `json:"resourceId,omitempty"`
	RessourceID      int       `json:"ressourceId,omitempty"`
	Title            string    `json:"title,omitempty"`
	Start            time.Time `json:"start"`
	End              time.Time `json:"end"`
	CheckinConfirmed bool      `json:"checkinConfirmed"`
}

func (b Booking) BookingID() int {
	if b.ID != 0 {
		return b.ID
	}
	return b.ReservationID
}

func (b Booking) MatchesResource(resourceID int) bool {
	return b.ResourceID == resourceID || b.RessourceID == resourceID || (b.ResourceID == 0 && b.RessourceID == 0)
}

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func (r TimeRange) Empty() bool {
	return !r.End.After(r.Start)
}

type Logger interface {
	Printf(format string, args ...any)
}
