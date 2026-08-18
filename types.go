package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	BusyRed         = "#FF0000"
	CheckinOrange   = "#FF6A00"
	FreeGreen       = "#00FF00"
	ProvisionalBlue = "#006DFF"
	UpcomingViolet  = "#7D00B3"
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
	case "#FF6A00":
		return CheckinOrange
	case "#00FF00":
		return FreeGreen
	case "#006DFF":
		return ProvisionalBlue
	case "#7D00B3":
		return UpcomingViolet
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

func (b *Booking) UnmarshalJSON(data []byte) error {
	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	b.ID = firstJSONInt(fields, "id", "Id", "ID")
	b.ReservationID = firstJSONInt(fields, "reservationId", "ReservationId", "ReservationID")
	b.ResourceID = firstJSONInt(fields, "resourceId", "ResourceId", "ResourceID")
	b.RessourceID = firstJSONInt(fields, "ressourceId", "RessourceId", "RessourceID")
	if b.ResourceID == 0 && b.RessourceID == 0 {
		b.RessourceID = nestedJSONInt(fields, "resource", "Resource", "ressource", "Ressource")
	}
	b.Title = firstJSONString(fields, "title", "Title")
	b.Start = firstJSONTime(fields, "start", "Start", "begin", "Begin", "beginn", "Beginn")
	b.End = firstJSONTime(fields, "end", "End", "ende", "Ende")
	b.CheckinConfirmed = firstJSONBool(fields, "checkinConfirmed", "CheckinConfirmed", "isCheckedIn", "IsCheckedIn")
	return nil
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

func firstJSONInt(fields map[string]json.RawMessage, names ...string) int {
	for _, name := range names {
		value, ok := fields[name]
		if !ok || string(value) == "null" {
			continue
		}
		var number int
		if err := json.Unmarshal(value, &number); err == nil {
			return number
		}
	}
	return 0
}

func nestedJSONInt(fields map[string]json.RawMessage, names ...string) int {
	for _, name := range names {
		value, ok := fields[name]
		if !ok || string(value) == "null" {
			continue
		}
		nested := map[string]json.RawMessage{}
		if err := json.Unmarshal(value, &nested); err != nil {
			continue
		}
		if id := firstJSONInt(nested, "id", "Id", "ID"); id != 0 {
			return id
		}
	}
	return 0
}

func firstJSONString(fields map[string]json.RawMessage, names ...string) string {
	for _, name := range names {
		value, ok := fields[name]
		if !ok || string(value) == "null" {
			continue
		}
		var text string
		if err := json.Unmarshal(value, &text); err == nil {
			return text
		}
	}
	return ""
}

func firstJSONBool(fields map[string]json.RawMessage, names ...string) bool {
	for _, name := range names {
		value, ok := fields[name]
		if !ok || string(value) == "null" {
			continue
		}
		var flag bool
		if err := json.Unmarshal(value, &flag); err == nil {
			return flag
		}
	}
	return false
}

func firstJSONTime(fields map[string]json.RawMessage, names ...string) time.Time {
	for _, name := range names {
		value, ok := fields[name]
		if !ok || string(value) == "null" {
			continue
		}
		var text string
		if err := json.Unmarshal(value, &text); err != nil || text == "" {
			continue
		}
		if parsed, ok := parseRoomsJSONTime(text); ok {
			return parsed
		}
	}
	return time.Time{}
}

func parseRoomsJSONTime(text string) (time.Time, bool) {
	if parsed, err := time.Parse(time.RFC3339Nano, text); err == nil {
		return parsed, true
	}
	for _, layout := range []string{
		"2006-01-02T15:04:05.9999999",
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05",
	} {
		if parsed, err := time.ParseInLocation(layout, text, time.UTC); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}
