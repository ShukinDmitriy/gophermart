package models

import (
	"fmt"
	"time"
)

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf(`%q`, time.Time(t).Format(time.RFC3339))
	return []byte(stamp), nil
}

func (t JSONTime) UnmarshalJSON(value []byte) error {
	_, err := time.Parse("\"2006-01-02T15:04:05Z\"", string(value))
	return err
}
