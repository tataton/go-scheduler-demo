package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go-scheduler-demo/models"

	"github.com/gin-gonic/gin"
)

// handlerGroup supplies a universal storage service
// (repo) to handler methods
type handlerGroup struct {
	repo storer
	sla  time.Duration
}

// unmarshalTimeSlot extracts an input TimeSlot from request
// payloads submitted to /availability routes. It makes assumptions about
// that request payload:
// 1. The payload should unmarshal to type TimeSlotJSON;
// 2. The Start JSON string property should be an RFC3339 timestamp that refers
//    to a time in the future; and
// 3. The Duration JSON string property should be a valid Duration string.
// Any failure to conform to these assumptions results in a BadRequest response
// to the API request.
func unmarshalTimeSlot(c *gin.Context) (models.TimeSlot, error) {
	// Unmarshal payload
	var timeSlotJSON models.TimeSlotJSON
	err := c.BindJSON(&timeSlotJSON)
	if err != nil {
		return models.TimeSlot{}, processBadRequestError(c, "request payload failed to marshal to TimeSlotJSON format")
	}
	inputTime, err := time.Parse(time.RFC3339, timeSlotJSON.Start)
	if err != nil {
		return models.TimeSlot{}, processBadRequestError(c, "request Start time could not be interpreted as an RFC3339 timestamp")
	}
	if inputTime.Before(time.Now()) {
		return models.TimeSlot{}, processBadRequestError(c, "request Start time cannot be missing or in the past")
	}
	inputDuration, err := time.ParseDuration(timeSlotJSON.Duration)
	if err != nil {
		return models.TimeSlot{}, processBadRequestError(c, "input Duration string could not be interpreted")
	}
	if inputDuration <= 0 {
		return models.TimeSlot{}, processBadRequestError(c, "input Duration cannot be missing, zero or negative")
	}
	timeSlot := models.TimeSlot{
		Start:    inputTime,
		Duration: inputDuration,
	}
	return timeSlot, nil
}

func processBadRequestError(c *gin.Context, errMsg string) error {
	err := errors.New(errMsg)
	// to caller:
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": errMsg})
	// to logger:
	c.Error(err)
	return err
}

// getAvailability accepts a request payload that defines
// a TimeSlot, and determines whether it is available--
// whether it overlaps wth any existing TimeSlot.
func (h *handlerGroup) getAvailability(c *gin.Context) {
	timeSlot, err := unmarshalTimeSlot(c)
	if err != nil {
		// response already prepared in unmarshalTimeSlot
		return
	}
	ctx, _ := context.WithTimeout(c.Request.Context(), h.sla)
	isUnavailable, err := h.repo.IsTimeSlotOverlapping(ctx, timeSlot)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"available": !isUnavailable})
}

func (h *handlerGroup) postAvailability(c *gin.Context) {
	timeSlot, err := unmarshalTimeSlot(c)
	if err != nil {
		// response already prepared in unmarshalTimeSlot
		return
	}
	ctx, _ := context.WithTimeout(c.Request.Context(), h.sla)
	isUnavailable, err := h.repo.IsTimeSlotOverlapping(ctx, timeSlot)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if isUnavailable {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"err": "some or all of requested time is already reserved"})
		return
	}
	err = h.repo.AddTimeSlot(c.Request.Context(), timeSlot)
	if err != nil {
		// to caller:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": "failed to add time slot"})
		// to logger:
		c.Error(err)
		return
	}
	c.Status(http.StatusCreated)
}

func (h *handlerGroup) deleteAvailability(c *gin.Context) {
	timeSlot, err := unmarshalTimeSlot(c)
	if err != nil {
		// response already prepared in unmarshalTimeSlot
		return
	}
	match, err := h.repo.IsTimeSlotExists(c.Request.Context(), timeSlot)
	if err != nil {
		// to caller:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": "failed to interrogate matching time slot"})
		// to logger:
		c.Error(err)
		return
	}
	if !match {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"err": "no matching time slot found"})
		return
	}
	err = h.repo.DeleteTimeSlot(c.Request.Context(), timeSlot)
	if err != nil {
		// to caller:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": "failed to delete matching time slot"})
		// to logger:
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
