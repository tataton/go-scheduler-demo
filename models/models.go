package models

import (
	"time"
)

type TimeSlotJSON struct {
	Start    string `json:"start"`
	Duration string `json:"duration"`
}

type TimeSlot struct {
	Start    time.Time
	Duration time.Duration
}
