package mapper

import (
	"errors"
	"fmt"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	genhandlers "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/server/http/handlers/generated"
	"github.com/google/uuid"
)

var ErrInvalidUUID = errors.New("invalid UUID format")

func CreateRequestToDomain(req genhandlers.CreateEventRequest) domain.Event {
	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	offsetTime := int64(0)
	if req.OffsetTime != nil {
		offsetTime = *req.OffsetTime
	}

	return domain.Event{
		Title:       req.Title,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Description: description,
		UserID:      req.UserId.String(),
		OffsetTime:  time.Duration(offsetTime) * time.Minute,
	}
}

func UpdateRequestToDomain(req genhandlers.UpdateEventRequest, id string) domain.Event {
	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	offsetTime := int64(0)
	if req.OffsetTime != nil {
		offsetTime = *req.OffsetTime
	}

	return domain.Event{
		ID:          id,
		Title:       req.Title,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Description: description,
		UserID:      req.UserId.String(),
		OffsetTime:  time.Duration(offsetTime) * time.Minute,
	}
}

func DomainToResponse(e domain.Event) (genhandlers.Event, error) {
	id, err := uuid.Parse(e.ID)
	if err != nil {
		return genhandlers.Event{}, fmt.Errorf("%w: %s", ErrInvalidUUID, e.ID)
	}

	userID, err := uuid.Parse(e.UserID)
	if err != nil {
		return genhandlers.Event{}, fmt.Errorf("%w: %s", ErrInvalidUUID, e.UserID)
	}

	offsetMinutes := int64(e.OffsetTime / time.Minute)

	return genhandlers.Event{
		Id:          &id,
		Title:       &e.Title,
		StartDate:   &e.StartDate,
		EndDate:     &e.EndDate,
		Description: &e.Description,
		UserId:      &userID,
		OffsetTime:  &offsetMinutes,
	}, nil
}

// DomainSliceToResponse converts slice of domain Events to slice of generated Events
func DomainSliceToResponse(events []domain.Event) ([]genhandlers.Event, error) {
	result := make([]genhandlers.Event, 0, len(events))
	for _, e := range events {
		event, err := DomainToResponse(e)
		if err != nil {
			return nil, err
		}
		result = append(result, event)
	}
	return result, nil
}
