package validators

import (
	"errors"
	"time"

	"go-scheduler-demo/models"
)

type JSONValidator struct{}

// ToTimeSlot enforces assumptions about a TimeSlot-containing request payload:
//
// 1. The payload should unmarshal to type TimeSlotJSON;
//
// 2. The Start JSON string property should be an RFC3339 timestamp that refers
// to a time in the future; and
//
// 3. The Duration JSON string property should be a valid Duration string.
//
// Any failure to conform to these assumptions results in a BadRequest response
// to the API request.
func (v *JSONValidator) ToTimeSlot(t models.TimeSlotJSON) (models.TimeSlot, error) {
	inputTime, err := time.Parse(time.RFC3339, t.Start)
	if err != nil {
		return models.TimeSlot{}, errors.New("request Start time could not be interpreted as an RFC3339 timestamp")
	}
	if inputTime.Before(time.Now()) {
		return models.TimeSlot{}, errors.New("request Start time cannot be missing or in the past")
	}
	inputDuration, err := time.ParseDuration(t.Duration)
	if err != nil {
		return models.TimeSlot{}, errors.New("input Duration string could not be interpreted")
	}
	if inputDuration <= 0 {
		return models.TimeSlot{}, errors.New("input Duration cannot be missing, zero or negative")
	}
	timeSlot := models.TimeSlot{
		Start:    inputTime,
		Duration: inputDuration,
	}
	return timeSlot, nil
}
