package server

//go:generate mockgen -source=./interface.go -package=server -destination=./mocks.go

import (
	"context"

	"go-scheduler-demo/models"
)

type storer interface {
	IsTimeSlotExists(ctx context.Context, timeSlot models.TimeSlot) (bool, error)
	IsTimeSlotOverlapping(ctx context.Context, timeSlot models.TimeSlot) (bool, error)
	AddTimeSlot(ctx context.Context, timeSlot models.TimeSlot) error
	DeleteTimeSlot(ctx context.Context, timeSlot models.TimeSlot) error
}
