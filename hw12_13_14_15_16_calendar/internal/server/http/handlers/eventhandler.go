package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/app"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	genhandlers "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/server/http/handlers/generated"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/server/http/handlers/mapper"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type EventHandler struct {
	app    app.Application
	logger logger.Logger
}

func NewEventHandler(app app.Application, log logger.Logger) *EventHandler {
	return &EventHandler{
		app:    app,
		logger: log,
	}
}

func (h *EventHandler) CreateEvent(ctx echo.Context) error {
	var req genhandlers.CreateEventRequest
	if err := ctx.Bind(&req); err != nil {
		h.logger.Error("failed to decode request: " + err.Error())
		return ctx.JSON(http.StatusBadRequest, genhandlers.ErrorResponse{Error: "invalid request body"})
	}

	event := mapper.CreateRequestToDomain(req)

	createdEvent, err := h.app.CreateEvent(ctx.Request().Context(), event)
	if err != nil {
		h.logger.Error("failed to create event: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, genhandlers.ErrorResponse{Error: "internal server error"})
	}

	h.logger.Info("event created successfully: " + createdEvent.ID)

	response, err := mapper.DomainToResponse(*createdEvent)
	if err != nil {
		h.logger.Error("failed to convert event to response: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, genhandlers.ErrorResponse{Error: "internal server error"})
	}
	return ctx.JSON(http.StatusCreated, response)
}

func (h *EventHandler) GetEvent(ctx echo.Context, id openapi_types.UUID) error {
	event, err := h.app.GetEventByID(ctx.Request().Context(), id.String())
	if err != nil {
		h.logger.Error("failed to get event: " + err.Error())
		return ctx.JSON(http.StatusNotFound, genhandlers.ErrorResponse{Error: "event not found"})
	}

	response, err := mapper.DomainToResponse(*event)
	if err != nil {
		h.logger.Error("failed to convert event to response: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, genhandlers.ErrorResponse{Error: "internal server error"})
	}
	return ctx.JSON(http.StatusOK, response)
}

func (h *EventHandler) UpdateEvent(ctx echo.Context, id openapi_types.UUID) error {
	var req genhandlers.UpdateEventRequest
	if err := ctx.Bind(&req); err != nil {
		h.logger.Error("failed to decode request: " + err.Error())
		return ctx.JSON(http.StatusBadRequest, genhandlers.ErrorResponse{Error: "invalid request body"})
	}

	event := mapper.UpdateRequestToDomain(req, id.String())

	updatedEvent, err := h.app.UpdateEvent(ctx.Request().Context(), id.String(), event)
	if err != nil {
		h.logger.Error("failed to update event: " + err.Error())
		if err.Error() == "event not found" {
			return ctx.JSON(http.StatusNotFound, genhandlers.ErrorResponse{Error: "event not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, genhandlers.ErrorResponse{Error: err.Error()})
	}

	h.logger.Info("event updated successfully: " + id.String())

	response, err := mapper.DomainToResponse(*updatedEvent)
	if err != nil {
		h.logger.Error("failed to convert event to response: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, genhandlers.ErrorResponse{Error: "internal server error"})
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h *EventHandler) DeleteEvent(ctx echo.Context, id openapi_types.UUID) error {
	if err := h.app.DeleteEvent(ctx.Request().Context(), id.String()); err != nil {
		h.logger.Error("failed to delete event: " + err.Error())

		if err.Error() == "event not found" {
			return ctx.JSON(http.StatusNotFound, genhandlers.ErrorResponse{Error: "event not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, genhandlers.ErrorResponse{Error: err.Error()})
	}

	h.logger.Info("event deleted successfully: " + id.String())
	return ctx.NoContent(http.StatusNoContent)
}

func (h *EventHandler) FindEvents(ctx echo.Context, params genhandlers.FindEventsParams) error {
	var userID string
	if params.UserId != nil {
		userID = params.UserId.String()
	}

	findedEvents, err := h.app.FindEvent(ctx.Request().Context(), userID, params.StartFrom, params.StartTo, params.EndFrom, params.EndTo)
	if err != nil {
		h.logger.Error("failed to find events: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, genhandlers.ErrorResponse{Error: err.Error()})
	}

	response, err := mapper.DomainSliceToResponse(findedEvents)
	if err != nil {
		h.logger.Error("failed to convert events to response: " + err.Error())
		return ctx.JSON(http.StatusInternalServerError, genhandlers.ErrorResponse{Error: "internal server error"})
	}

	return ctx.JSON(http.StatusOK, response)
}
