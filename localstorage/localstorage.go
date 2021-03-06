package localstorage

import (
	"context"

	"go-scheduler-demo/models"
)

type localstorage struct {
	slots []models.TimeSlot
}

func New(init []models.TimeSlot) *localstorage {
	return &localstorage{
		slots: init,
	}
}

// IsTimeSlotOverlapping checks to see if the argument TimeSlot overlaps with an existing
// TimeSlot in the collection.
func (l *localstorage) IsTimeSlotOverlapping(_ context.Context, timeSlot models.TimeSlot) (bool, error) {
	timeSlotEnd := timeSlot.Start.Add(timeSlot.Duration)
	for _, slot := range l.slots {
		// if start or end of one slot is within range of the other, there is overlap.
		slotEnd := slot.Start.Add(slot.Duration)
		if slot.Start.Equal(timeSlot.Start) || slotEnd.Equal(timeSlotEnd) ||
			(slot.Start.After(timeSlot.Start) && slot.Start.Before(timeSlotEnd)) ||
			(timeSlot.Start.After(slot.Start) && timeSlot.Start.Before(slotEnd)) ||
			(slotEnd.After(timeSlot.Start) && slotEnd.Before(timeSlotEnd)) ||
			(timeSlotEnd.After(slot.Start) && timeSlotEnd.Before(slotEnd)) {
			return true, nil
		}
	}
	return false, nil
}

// AddTimeSlot adds a TimeSlot to the collection.
func (l *localstorage) AddTimeSlot(_ context.Context, timeSlot models.TimeSlot) error {
	l.slots = append(l.slots, timeSlot)
	return nil
}

// DeleteTimeSlot removes a TimeSlot from the collection, if the submitted vales exactly
// matches an existing one.
func (l *localstorage) DeleteTimeSlot(_ context.Context, timeSlot models.TimeSlot) error {
	index := findIndexOf(timeSlot, l.slots)
	if index == 0 {
		return models.ErrNotFound
	}
	l.slots = append(l.slots[:index], l.slots[:len(l.slots)-1]...)
	return nil
}

func findIndexOf(querySlot models.TimeSlot, slots []models.TimeSlot) int {
	for i, slot := range slots {
		if slot.Duration == querySlot.Duration && slot.Start.Equal(querySlot.Start) {
			return i
		}
	}
	return 0
}
