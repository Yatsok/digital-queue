package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/Yatsok/digital-queue/internal/controllers"
	"github.com/Yatsok/digital-queue/internal/models"
	"github.com/Yatsok/digital-queue/templates"
	"github.com/angelofallars/htmx-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type EventHandler struct {
	eventStore controllers.EventStore
	slotStore  controllers.SlotStore
	userStore  controllers.UserStore
	authStore  controllers.AuthStore
	CustomErrorHandler
}

const (
	EventsPerPage = 10
)

func NewEventHandler(userStore controllers.UserStore, slotStore controllers.SlotStore, eventStore controllers.EventStore, authStore controllers.AuthStore) *EventHandler {
	return &EventHandler{
		eventStore: eventStore,
		slotStore:  slotStore,
		userStore:  userStore,
		authStore:  authStore,
	}
}

// GET Request on "/event/new"
// Return to "/" if not logged
// Render templ component CreateEventForm
func (h *EventHandler) CreateEventFormHandler(w http.ResponseWriter, r *http.Request) {
	req := &models.EventRequestGET{
		AuthStatus: h.authStore.GetAuthStatusContext(w, r),
	}

	component := templates.Base(templates.CreateEventForm(req))
	component.Render(r.Context(), w)
}

// GET Request on "/event/upload"
// Return to "/" if not logged
// Render templ component UploadImageForm
func (h *EventHandler) UploadImageFormHandler(w http.ResponseWriter, r *http.Request) {
	event, err := h.eventStore.GetEventByVars(r)
	if err != nil {
		return
	}

	req := &models.EventRequestGET{
		AuthStatus: h.authStore.GetAuthStatusContext(w, r),
		Event:      *event,
	}

	if htmx.IsHTMX(r) {
		w.Header().Set("HX-Redirect", fmt.Sprintf("/event/%s/upload", event.ID))
	}

	component := templates.Base(templates.UploadImageForm(req))
	component.Render(r.Context(), w)
}

// POST Request on "/event/new"
// Creates new event in DB
// Redirects to СreateSlots endpoint ("/event/add")
func (h *EventHandler) CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	var req models.EventRequestPOST
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	event := models.NewEvent(
		h.authStore.GetUserID(r),
		req.Name,
		req.Description,
	)

	err := h.eventStore.SaveEvent(event)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/event/%s/upload", event.ID))
}

// GET Request on "/event/edit/{id}"
// Return to "/" if not logged
// Render templ component EditEventForm
func (h *EventHandler) EditEventFormHandler(w http.ResponseWriter, r *http.Request) {
	userID := h.authStore.GetUserID(r)

	event, err := h.eventStore.GetEventByVars(r)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	ownershipStatus := h.eventStore.CheckOwnership(userID, event.UserID)
	h.eventStore.OwnershipRedirect(w, r, ownershipStatus)

	req := &models.EventRequestGET{
		AuthStatus: h.authStore.GetAuthStatusContext(w, r),
		Event:      *event,
	}

	if htmx.IsHTMX(r) {
		w.Header().Set("HX-Redirect", fmt.Sprintf("/event/%s/edit", event.ID))
	}

	component := templates.Base(templates.EditEventForm(req))
	component.Render(r.Context(), w)
}

// PUT Request on "/event/edit"
// Updates event in DB
// Redirects to events details endpoint ("/event/{id}")
func (h *EventHandler) EditEventHandler(w http.ResponseWriter, r *http.Request) {
	var req models.EventRequestPOST
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	event, err := h.eventStore.GetEventByID(uuid.MustParse(req.ID))
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	userID := h.authStore.GetUserID(r)
	ownershipStatus := h.eventStore.CheckOwnership(userID, event.UserID)
	h.eventStore.OwnershipRedirect(w, r, ownershipStatus)

	err = h.eventStore.UpdateEvent(
		uuid.MustParse(req.ID),
		req.Name,
		req.Description,
	)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/event/%s", uuid.MustParse(req.ID)))
}

// GET Request on "/event/{id}"
// Checks if the user is the event's creator, if the user is the creator, show editiing tools
// Render templ component EventDetails
func (h *EventHandler) EventDetailsHandler(w http.ResponseWriter, r *http.Request) {
	event, err := h.eventStore.GetEventByVars(r)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	if htmx.IsHTMX(r) {
		w.Header().Set("HX-Redirect", fmt.Sprintf("/event/%s", event.ID))
	}

	userID := h.authStore.GetUserID(r)

	timeSlots, err := h.slotStore.GetTimeSlotsByEventID(event.ID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	availableTimeSlots := map[string][]models.TimeSlot{
		event.ID.String(): timeSlots,
	}

	req := &models.EventRequestGET{
		AuthStatus:      h.authStore.GetAuthStatusContext(w, r),
		Event:           *event,
		UserID:          userID,
		OwnershipStatus: h.eventStore.CheckOwnership(userID, event.UserID),
		TimeSlots:       h.slotStore.GetFilteredTimeSlots(availableTimeSlots, userID),
		Timezone:        h.userStore.GetUserTimezone(userID),
	}

	component := templates.Base(templates.EventDetails(req))
	component.Render(r.Context(), w)
}

// GET Request on "/event/{id}/slots"
// Checks if the user is the event's creator, if the user is the creator, show editiing tools
// Render templ component EventDetails
func (h *EventHandler) EventSlotsHandler(w http.ResponseWriter, r *http.Request) {
	event, err := h.eventStore.GetEventByVars(r)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	userID := h.authStore.GetUserID(r)
	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		return
	}

	timeSlots, err := h.slotStore.GetTimeSlotsByEventID(event.ID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	availableTimeSlots := map[string][]models.TimeSlot{
		event.ID.String(): timeSlots,
	}

	req := &models.EventRequestGET{
		UserID:          userID,
		Event:           *event,
		OwnershipStatus: h.eventStore.CheckOwnership(userID, event.UserID),
		TimeSlots:       h.slotStore.GetFilteredTimeSlots(availableTimeSlots, userID),
		Timezone:        h.userStore.GetUserTimezone(userID),
		ImagePath:       user.ImagePath,
	}

	component := templates.TimeSlotTable(req)
	component.Render(r.Context(), w)
}

// GET Request on "/events/{page}"
// Getes all events from DB
// Render templ component EventsList
func (h *EventHandler) EventsListHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, err := strconv.Atoi(vars["page"])
	if err != nil || page < 1 {
		page = 1
	}

	offset := (page - 1) * EventsPerPage

	events, err := h.eventStore.GetPaginatedEvents(offset, EventsPerPage)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	totalEvents, err := h.eventStore.CountAllEvents()
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}
	totalPages := int(math.Ceil(float64(totalEvents) / float64(EventsPerPage)))
	pagination := h.eventStore.CalculatePagination(page, totalPages)

	allTimeSlots := make(map[string][]models.TimeSlot)
	for _, event := range events {
		timeSlots, err := h.slotStore.GetTimeSlotsByEventID(event.ID)
		if err != nil {
			h.CustomErrorHandler.InternalServerErrorHandler(w, r)
			return
		}
		allTimeSlots[event.ID.String()] = timeSlots
	}

	if htmx.IsHTMX(r) {
		w.Header().Set("HX-Redirect", fmt.Sprintf("/events/%d", page))
	}

	req := &models.EventRequestGET{
		AuthStatus: h.authStore.GetAuthStatusContext(w, r),
		Events:     events,
		Pagination: pagination,
		Page:       page,
		TotalPages: totalPages,
		TimeSlots:  allTimeSlots,
	}

	component := templates.Base(templates.EventsList(req))
	component.Render(r.Context(), w)
}

// DELETE Request on "/events/{id}/delete"
// Expected behaviour: delete the event with all its associated time slots from DB
func (h *EventHandler) DeleteEventHandler(w http.ResponseWriter, r *http.Request) {
	userID := h.authStore.GetUserID(r)

	event, err := h.eventStore.GetEventByVars(r)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	ownershipStatus := h.eventStore.CheckOwnership(userID, event.UserID)
	h.eventStore.OwnershipRedirect(w, r, ownershipStatus)

	err = h.eventStore.DeleteEvent(event.ID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	w.Header().Set("HX-Redirect", "/events")
}
