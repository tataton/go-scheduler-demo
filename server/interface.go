package server

//go:generate mockgen -source=./interface.go -package=server -destination=./mocks.go

import (
	"context"

	"go-scheduler-demo/models"
)

type storer interface {
	IsTimeSlotOverlapping(ctx context.Context, timeSlot models.TimeSlot) (bool, error)
	AddTimeSlot(ctx context.Context, timeSlot models.TimeSlot) error
	DeleteTimeSlot(ctx context.Context, timeSlot models.TimeSlot) error
}

type jsonValidator interface {
	ToTimeSlot(t models.TimeSlotJSON) (models.TimeSlot, error)
}
