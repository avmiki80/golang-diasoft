package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/app"
	events "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/gorilla/mux"
)

type EventHandler struct {
	app    app.Application
	logger logger.Logger
}

type CreateEventRequest struct {
	Title       string `json:"title"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Description string `json:"description"`
	UserID      string `json:"userId"`
	OffsetTime  int64  `json:"offsetTime"`
}

type UpdateEventRequest struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Description string `json:"description"`
	UserID      string `json:"userId"`
	OffsetTime  int64  `json:"offsetTime"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewEventHandler(app app.Application, log logger.Logger) *EventHandler {
	return &EventHandler{
		app:    app,
		logger: log,
	}
}

func (h *EventHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/events", h.CreateEvent).Methods(http.MethodPost)
	router.HandleFunc("/events/{id}", h.GetEvent).Methods(http.MethodGet)
	router.HandleFunc("/events/{id}", h.UpdateEvent).Methods(http.MethodPut)
	router.HandleFunc("/events/{id}", h.DeleteEvent).Methods(http.MethodDelete)
	router.HandleFunc("/events", h.FindEvents).Methods(http.MethodGet)
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request: " + err.Error())
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		h.logger.Error("failed to parse start_date: " + err.Error())
		h.respondError(w, http.StatusBadRequest, "invalid start_date format, use RFC3339")
		return
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		h.logger.Error("failed to parse end_date: " + err.Error())
		h.respondError(w, http.StatusBadRequest, "invalid end_date format, use RFC3339")
		return
	}

	event := events.Event{
		Title:       req.Title,
		StartDate:   startDate,
		EndDate:     endDate,
		Description: req.Description,
		UserID:      req.UserID,
		OffsetTime:  time.Duration(req.OffsetTime) * time.Minute,
	}
	createdEvent, err := h.app.CreateEvent(r.Context(), event)
	if err != nil {
		h.logger.Error("failed to create event: " + err.Error())
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("event created successfully: " + createdEvent.ID)
	h.respondSuccess(w, http.StatusCreated, "", createdEvent)
}

func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	event, err := h.app.GetEventByID(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get event: " + err.Error())
		h.respondError(w, http.StatusNotFound, "event not found")
		return
	}

	h.respondSuccess(w, http.StatusOK, "", event)
}

func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request: " + err.Error())
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		h.logger.Error("failed to parse start_date: " + err.Error())
		h.respondError(w, http.StatusBadRequest, "invalid start_date format, use RFC3339")
		return
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		h.logger.Error("failed to parse end_date: " + err.Error())
		h.respondError(w, http.StatusBadRequest, "invalid end_date format, use RFC3339")
		return
	}

	event := events.Event{
		ID:          id,
		Title:       req.Title,
		StartDate:   startDate,
		EndDate:     endDate,
		Description: req.Description,
		UserID:      req.UserID,
		OffsetTime:  time.Duration(req.OffsetTime) * time.Minute,
	}
	updatedEvent, err := h.app.UpdateEvent(r.Context(), id, event)
	if err != nil {
		h.logger.Error("failed to update event: " + err.Error())
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("event updated successfully: " + id)
	h.respondSuccess(w, http.StatusOK, "", updatedEvent)
}

func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.app.DeleteEvent(r.Context(), id); err != nil {
		h.logger.Error("failed to delete event: " + err.Error())
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("event deleted successfully: " + id)
	h.respondSuccess(w, http.StatusOK, "event deleted successfully", id)
}

func (h *EventHandler) FindEvents(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	startFromStr := r.URL.Query().Get("startFrom")
	startToStr := r.URL.Query().Get("startTo")
	endFromStr := r.URL.Query().Get("endFrom")
	endToStr := r.URL.Query().Get("endTo")

	var startFrom, startTo, endFrom, endTo *time.Time

	if startFromStr != "" {
		t, err := time.Parse(time.RFC3339, startFromStr)
		if err != nil {
			h.logger.Error("failed to parse start_from: " + err.Error())
			h.respondError(w, http.StatusBadRequest, "invalid start_from format, use RFC3339")
			return
		}
		startFrom = &t
	}

	if startToStr != "" {
		t, err := time.Parse(time.RFC3339, startToStr)
		if err != nil {
			h.logger.Error("failed to parse start_to: " + err.Error())
			h.respondError(w, http.StatusBadRequest, "invalid start_to format, use RFC3339")
			return
		}
		startTo = &t
	}

	if endFromStr != "" {
		t, err := time.Parse(time.RFC3339, endFromStr)
		if err != nil {
			h.logger.Error("failed to parse end_from: " + err.Error())
			h.respondError(w, http.StatusBadRequest, "invalid end_from format, use RFC3339")
			return
		}
		endFrom = &t
	}

	if endToStr != "" {
		t, err := time.Parse(time.RFC3339, endToStr)
		if err != nil {
			h.logger.Error("failed to parse end_to: " + err.Error())
			h.respondError(w, http.StatusBadRequest, "invalid end_to format, use RFC3339")
			return
		}
		endTo = &t
	}

	findedEvents, err := h.app.FindEvent(r.Context(), userID, startFrom, startTo, endFrom, endTo)
	if err != nil {
		h.logger.Error("failed to find events: " + err.Error())
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondSuccess(w, http.StatusOK, "", findedEvents)
}

func (h *EventHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(ErrorResponse{Error: message})
	if err != nil {
		h.logger.Error("failed to encode ErrorResponse: " + err.Error())
	}
}

func (h *EventHandler) respondSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(SuccessResponse{Message: message, Data: data})
	if err != nil {
		h.logger.Error("failed to encode SuccessResponse: " + err.Error())
	}
}
