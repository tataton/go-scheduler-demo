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
	validator jsonValidator
	repo      storer
	sla       time.Duration
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
	// Unmarshal payload
	var timeSlotJSON models.TimeSlotJSON
	err := c.BindJSON(&timeSlotJSON) // built-in analog to encoding/json.Unmarshal
	if err != nil {
		errMsg := "request payload failed to marshal to TimeSlotJSON format"
		// to caller:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": errMsg})
		// to logger:
		c.Error(err)
		return
	}
	timeSlot, err := h.validator.ToTimeSlot(timeSlotJSON)
	if err != nil {
		// to caller:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		// to logger:
		c.Error(err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.sla)
	defer cancel()
	isUnavailable, err := h.repo.IsTimeSlotOverlapping(ctx, timeSlot)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"available": !isUnavailable})
}

// getAvailability accepts a request payload that defines
// a TimeSlot, determines whether it overlaps an existing one,
// and if not, calls the stroage dependency to add it.
func (h *handlerGroup) postAvailability(c *gin.Context) {
	// Unmarshal payload
	var timeSlotJSON models.TimeSlotJSON
	err := c.BindJSON(&timeSlotJSON)
	if err != nil {
		errMsg := "request payload failed to marshal to TimeSlotJSON format"
		// to caller:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": errMsg})
		// to logger:
		c.Error(err)
		return
	}
	timeSlot, err := h.validator.ToTimeSlot(timeSlotJSON)
	if err != nil {
		// to caller:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		// to logger:
		c.Error(err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.sla)
	defer cancel()
	// Here, I'm having this handler execute a sequence of business logic:
	//
	// 1. Check to see if the submitted timeSlot overlaps with an existing one;
	// and
	//
	// 2. Add the timeSlot.
	//
	// I wouldn't normally do that in a handler, whose typical domain of
	// responsibility is request and response handling. But this is the only
	// one of the three that has independent business logic ("we should not
	// create overlapping timeSlots"), so I'll just execute it directly here.
	isUnavailable, err := h.repo.IsTimeSlotOverlapping(ctx, timeSlot)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if isUnavailable {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"err": "some or all of requested time is already reserved"})
		return
	}
	err = h.repo.AddTimeSlot(ctx, timeSlot)
	if err != nil {
		// to caller:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": "failed to add time slot"})
		// to logger:
		c.Error(err)
		return
	}
	c.Status(http.StatusCreated)
}

// deleteAvailability accepts a request payload that defines
// a TimeSlot, and removes the timeSlot if it exactly matches
// an existing one.
func (h *handlerGroup) deleteAvailability(c *gin.Context) {
	// Unmarshal payload
	var timeSlotJSON models.TimeSlotJSON
	err := c.BindJSON(&timeSlotJSON)
	if err != nil {
		errMsg := "request payload failed to marshal to TimeSlotJSON format"
		// to caller:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": errMsg})
		// to logger:
		c.Error(err)
		return
	}
	timeSlot, err := h.validator.ToTimeSlot(timeSlotJSON)
	if err != nil {
		// to caller:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		// to logger:
		c.Error(err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.sla)
	defer cancel()
	err = h.repo.DeleteTimeSlot(ctx, timeSlot)
	if err == models.ErrNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"err": "no matching time slot found"})
		return
	}
	if err != nil {
		// to caller:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": "failed to delete matching time slot"})
		// to logger:
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
