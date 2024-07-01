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
