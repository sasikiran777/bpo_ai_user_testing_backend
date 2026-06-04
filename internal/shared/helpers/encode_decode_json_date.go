package helpers

import (
	"encoding/json"
	"strings"
	"time"
)

type FlexibleDateTime struct {
	time.Time
}

func (f *FlexibleDateTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		f.Time = time.Time{}
		return nil
	}
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"02-01-2006", // DD-MM-YYYY, e.g. 30-05-2027
		"2006-01-02", // YYYY-MM-DD
	}
	var lastErr error
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, s, time.UTC)
		if err == nil {
			f.Time = t
			return nil
		}
		lastErr = err
	}
	return lastErr
}

func (f FlexibleDateTime) MarshalJSON() ([]byte, error) {
	if f.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(f.Time.Format(time.RFC3339Nano))
}
