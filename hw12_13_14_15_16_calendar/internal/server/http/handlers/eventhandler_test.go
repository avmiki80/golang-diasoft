package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/domain"
	genhandlers "github.com/avmiki80/golang-diasoft/hw12_13_14_15_16_calendar/internal/server/http/handlers/generated"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEventHandler_CreateEvent_Success(t *testing.T) {
	mockApp := new(MockApplication)
	mockLogger := new(MockLogger)
	handler := NewEventHandler(mockApp, mockLogger)

	userID := uuid.New()
	eventID := uuid.New().String()

	createdEvent := &domain.Event{
		ID:          eventID,
		Title:       "Test Event",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "Test Description",
		UserID:      userID.String(),
		OffsetTime:  0,
	}

	mockApp.On("CreateEvent", mock.Anything, mock.MatchedBy(func(e domain.Event) bool {
		return e.Title == "Test Event" && e.UserID == userID.String()
	})).Return(createdEvent, nil)
	mockLogger.On("Info", mock.Anything).Return()

	e := echo.New()
	reqBody := `{
		"title": "Test Event",
		"startDate": "2024-01-01T10:00:00Z",
		"endDate": "2024-01-01T11:00:00Z",
		"description": "Test Description",
		"userId": "` + userID.String() + `",
		"offsetTime": 0
	}`
	req := httptest.NewRequest(http.MethodPost, "/event", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreateEvent(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response genhandlers.Event
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Test Event", *response.Title)
	assert.Equal(t, userID, *response.UserId)

	mockApp.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestEventHandler_CreateEvent_InvalidRequest(t *testing.T) {
	mockApp := new(MockApplication)
	mockLogger := new(MockLogger)
	handler := NewEventHandler(mockApp, mockLogger)

	mockLogger.On("Error", mock.Anything).Return()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/event", strings.NewReader("invalid json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreateEvent(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response genhandlers.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid request body", response.Error)

	mockLogger.AssertExpectations(t)
}

func TestEventHandler_CreateEvent_ServiceError(t *testing.T) {
	mockApp := new(MockApplication)
	mockLogger := new(MockLogger)
	handler := NewEventHandler(mockApp, mockLogger)

	userID := uuid.New()

	mockApp.On("CreateEvent", mock.Anything, mock.Anything).Return(nil, errors.New("service error"))
	mockLogger.On("Error", mock.Anything).Return()

	e := echo.New()
	reqBody := `{
		"title": "Test Event",
		"startDate": "2024-01-01T10:00:00Z",
		"endDate": "2024-01-01T11:00:00Z",
		"userId": "` + userID.String() + `"
	}`
	req := httptest.NewRequest(http.MethodPost, "/event", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreateEvent(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response genhandlers.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "internal server error", response.Error)

	mockApp.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestEventHandler_GetEvent_Success(t *testing.T) {
	mockApp := new(MockApplication)
	mockLogger := new(MockLogger)
	handler := NewEventHandler(mockApp, mockLogger)

	eventID := uuid.New()
	userID := uuid.New()

	event := &domain.Event{
		ID:          eventID.String(),
		Title:       "Test Event",
		StartDate:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Description: "Test Description",
		UserID:      userID.String(),
		OffsetTime:  0,
	}

	mockApp.On("GetEventByID", mock.Anything, eventID.String()).Return(event, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/event/"+eventID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetEvent(c, eventID)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response genhandlers.Event
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Test Event", *response.Title)
	assert.Equal(t, eventID, *response.Id)

	mockApp.AssertExpectations(t)
}

func TestEventHandler_GetEvent_NotFound(t *testing.T) {
	mockApp := new(MockApplication)
	mockLogger := new(MockLogger)
	handler := NewEventHandler(mockApp, mockLogger)

	eventID := uuid.New()

	mockApp.On("GetEventByID", mock.Anything, eventID.String()).Return(nil, errors.New("event not found"))
	mockLogger.On("Error", mock.Anything).Return()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/event/"+eventID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetEvent(c, eventID)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response genhandlers.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "event not found", response.Error)

	mockApp.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestEventHandler_UpdateEvent_Success(t *testing.T) {
	mockApp := new(MockApplication)
	mockLogger := new(MockLogger)
	handler := NewEventHandler(mockApp, mockLogger)

	eventID := uuid.New()
	userID := uuid.New()

	updatedEvent := &domain.Event{
		ID:          eventID.String(),
		Title:       "Updated Title",
		StartDate:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		Description: "Updated Description",
		UserID:      userID.String(),
		OffsetTime:  30 * time.Minute,
	}

	mockApp.On("UpdateEvent", mock.Anything, eventID.String(), mock.Anything).Return(updatedEvent, nil)
	mockLogger.On("Info", mock.Anything).Return()

	e := echo.New()
	reqBody := `{
		"title": "Updated Title",
		"startDate": "2024-01-01T12:00:00Z",
		"endDate": "2024-01-01T13:00:00Z",
		"description": "Updated Description",
		"userId": "` + userID.String() + `",
		"offsetTime": 30
	}`
	req := httptest.NewRequest(http.MethodPut, "/event/"+eventID.String(), strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.UpdateEvent(c, eventID)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response genhandlers.Event
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", *response.Title)
	assert.Equal(t, "Updated Description", *response.Description)

	mockApp.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestEventHandler_UpdateEvent_NotFound(t *testing.T) {
	mockApp := new(MockApplication)
	mockLogger := new(MockLogger)
	handler := NewEventHandler(mockApp, mockLogger)

	eventID := uuid.New()
	userID := uuid.New()

	mockApp.On("UpdateEvent", mock.Anything, eventID.String(), mock.Anything).Return(nil, errors.New("event not found"))
	mockLogger.On("Error", mock.Anything).Return()

	e := echo.New()
	reqBody := `{
		"title": "Test Event",
		"startDate": "2024-01-01T10:00:00Z",
		"endDate": "2024-01-01T11:00:00Z",
		"userId": "` + userID.String() + `"
	}`
	req := httptest.NewRequest(http.MethodPut, "/event/"+eventID.String(), strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.UpdateEvent(c, eventID)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response genhandlers.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "event not found", response.Error)

	mockApp.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestEventHandler_UpdateEvent_InvalidRequest(t *testing.T) {
	mockApp := new(MockApplication)
	mockLogger := new(MockLogger)
	handler := NewEventHandler(mockApp, mockLogger)

	eventID := uuid.New()

	mockLogger.On("Error", mock.Anything).Return()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/event/"+eventID.String(), strings.NewReader("invalid json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.UpdateEvent(c, eventID)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response genhandlers.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid request body", response.Error)

	mockLogger.AssertExpectations(t)
}

func TestEventHandler_DeleteEvent_Success(t *testing.T) {
	mockApp := new(MockApplication)
	mockLogger := new(MockLogger)
	handler := NewEventHandler(mockApp, mockLogger)

	eventID := uuid.New()

	mockApp.On("DeleteEvent", mock.Anything, eventID.String()).Return(nil)
	mockLogger.On("Info", mock.Anything).Return()

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/event/"+eventID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.DeleteEvent(c, eventID)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	mockApp.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestEventHandler_DeleteEvent_NotFound(t *testing.T) {
	mockApp := new(MockApplication)
	mockLogger := new(MockLogger)
	handler := NewEventHandler(mockApp, mockLogger)

	eventID := uuid.New()

	mockApp.On("DeleteEvent", mock.Anything, eventID.String()).Return(errors.New("event not found"))
	mockLogger.On("Error", mock.Anything).Return()

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/event/"+eventID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.DeleteEvent(c, eventID)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response genhandlers.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "event not found", response.Error)

	mockApp.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestEventHandler_FindEvents_ByUserID(t *testing.T) {
	mockApp := new(MockApplication)
	mockLogger := new(MockLogger)
	handler := NewEventHandler(mockApp, mockLogger)

	userID := uuid.New()

	events := []domain.Event{
		{
			ID:         uuid.New().String(),
			Title:      "Event 1",
			StartDate:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			UserID:     userID.String(),
			OffsetTime: 0,
		},
		{
			ID:         uuid.New().String(),
			Title:      "Event 2",
			StartDate:  time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC),
			UserID:     userID.String(),
			OffsetTime: 0,
		},
	}

	mockApp.On("FindEvent", mock.Anything, userID.String(), (*time.Time)(nil), (*time.Time)(nil), (*time.Time)(nil), (*time.Time)(nil)).Return(events, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/event?userId="+userID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := genhandlers.FindEventsParams{
		UserId: &userID,
	}

	err := handler.FindEvents(c, params)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response []genhandlers.Event
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "Event 1", *response[0].Title)
	assert.Equal(t, "Event 2", *response[1].Title)

	mockApp.AssertExpectations(t)
}

func TestEventHandler_FindEvents_ByDateRange(t *testing.T) {
	mockApp := new(MockApplication)
	mockLogger := new(MockLogger)
	handler := NewEventHandler(mockApp, mockLogger)

	userID := uuid.New()
	startFrom := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	startTo := time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC)

	events := []domain.Event{
		{
			ID:         uuid.New().String(),
			Title:      "Event Jan 5",
			StartDate:  time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2024, 1, 5, 11, 0, 0, 0, time.UTC),
			UserID:     userID.String(),
			OffsetTime: 0,
		},
	}

	mockApp.On("FindEvent", mock.Anything, userID.String(), &startFrom, &startTo, (*time.Time)(nil), (*time.Time)(nil)).Return(events, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/event", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := genhandlers.FindEventsParams{
		UserId:    &userID,
		StartFrom: &startFrom,
		StartTo:   &startTo,
	}

	err := handler.FindEvents(c, params)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response []genhandlers.Event
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, "Event Jan 5", *response[0].Title)

	mockApp.AssertExpectations(t)
}

func TestEventHandler_FindEvents_AllEvents(t *testing.T) {
	mockApp := new(MockApplication)
	mockLogger := new(MockLogger)
	handler := NewEventHandler(mockApp, mockLogger)

	events := []domain.Event{
		{
			ID:         uuid.New().String(),
			Title:      "Event A",
			StartDate:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			UserID:     uuid.New().String(),
			OffsetTime: 0,
		},
		{
			ID:         uuid.New().String(),
			Title:      "Event B",
			StartDate:  time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC),
			UserID:     uuid.New().String(),
			OffsetTime: 0,
		},
	}

	mockApp.On("FindEvent", mock.Anything, "", (*time.Time)(nil), (*time.Time)(nil), (*time.Time)(nil), (*time.Time)(nil)).Return(events, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/event", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := genhandlers.FindEventsParams{}

	err := handler.FindEvents(c, params)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response []genhandlers.Event
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(response), 2)

	mockApp.AssertExpectations(t)
}

func TestEventHandler_FindEvents_ServiceError(t *testing.T) {
	mockApp := new(MockApplication)
	mockLogger := new(MockLogger)
	handler := NewEventHandler(mockApp, mockLogger)

	mockApp.On("FindEvent", mock.Anything, "", (*time.Time)(nil), (*time.Time)(nil), (*time.Time)(nil), (*time.Time)(nil)).Return(nil, errors.New("database error"))
	mockLogger.On("Error", mock.Anything).Return()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/event", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := genhandlers.FindEventsParams{}

	err := handler.FindEvents(c, params)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response genhandlers.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "database error", response.Error)

	mockApp.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestEventHandler_CreateEvent_InvalidUUID(t *testing.T) {
	mockApp := new(MockApplication)
	mockLogger := new(MockLogger)
	handler := NewEventHandler(mockApp, mockLogger)

	mockLogger.On("Error", mock.Anything).Return()

	e := echo.New()
	// Невалидный UUID в userId приведет к ошибке парсинга при Bind
	reqBody := `{
		"title": "Test Event",
		"startDate": "2024-01-01T10:00:00Z",
		"endDate": "2024-01-01T11:00:00Z",
		"description": "Test Description",
		"userId": "invalid-uuid",
		"offsetTime": 0
	}`
	req := httptest.NewRequest(http.MethodPost, "/event", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreateEvent(c)

	require.NoError(t, err)
	// Ошибка парсинга UUID при Bind -> BadRequest
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response genhandlers.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "invalid request body", response.Error)

	mockLogger.AssertExpectations(t)
}
