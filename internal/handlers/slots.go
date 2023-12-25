package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Yatsok/digital-queue/internal/controllers"
	"github.com/Yatsok/digital-queue/internal/helper"
	"github.com/Yatsok/digital-queue/internal/models"
	"github.com/Yatsok/digital-queue/templates"
	"github.com/angelofallars/htmx-go"
	"github.com/google/uuid"
)

type SlotHandler struct {
	slotStore  controllers.SlotStore
	eventStore controllers.EventStore
	userStore  controllers.UserStore
	authStore  controllers.AuthStore
	CustomErrorHandler
}

func NewSlotHandler(userStore controllers.UserStore, slotStore controllers.SlotStore, eventStore controllers.EventStore, authStore controllers.AuthStore) *SlotHandler {
	return &SlotHandler{
		slotStore:  slotStore,
		userStore:  userStore,
		eventStore: eventStore,
		authStore:  authStore,
	}
}

// GET Request on "/event/{id}/add"
// Return to "/" if not logged
// Render templ component CreateTimeSlots
func (h *SlotHandler) CreateTimeSlotsFormHandler(w http.ResponseWriter, r *http.Request) {
	userID := h.authStore.GetUserID(r)

	event, err := h.eventStore.GetEventByVars(r)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	ownershipStatus := h.eventStore.CheckOwnership(userID, event.UserID)
	h.eventStore.OwnershipRedirect(w, r, ownershipStatus)

	timeSlots, err := h.slotStore.GetTimeSlotsByEventID(event.ID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	if htmx.IsHTMX(r) {
		w.Header().Set("HX-Redirect", fmt.Sprintf("/event/%s/add", event.ID))
	}

	req := models.SlotRequestGET{
		AuthStatus:      h.authStore.GetAuthStatusContext(w, r),
		OwnershipStatus: ownershipStatus,
		Event:           *event,
		UserID:          userID,
		TimeSlots:       timeSlots,
		Timezone:        h.userStore.GetUserTimezone(userID),
	}

	component := templates.Base(templates.CreateTimeSlots(req))
	component.Render(r.Context(), w)
}

// POST Request on "/event/add"
// Updates event in DB
// Redirects to events details endpoint ("/event/{id}")
func (h *SlotHandler) CreateTimeSlotHandler(w http.ResponseWriter, r *http.Request) {
	var req models.SlotRequestPOST
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	userID := h.authStore.GetUserID(r)

	timezone := h.userStore.GetUserTimezone(userID)

	event, err := h.eventStore.GetEventByID(uuid.MustParse(req.EventID))
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	ownershipStatus := h.eventStore.CheckOwnership(userID, event.UserID)
	h.eventStore.OwnershipRedirect(w, r, ownershipStatus)

	var nilUserID *uuid.UUID

	timeSlot := models.NewTimeSlot(
		nilUserID,
		event.ID,
		"",
		event.Name,
		helper.TimeInLocation(req.StartTime, timezone).UTC(),
		helper.TimeInLocation(req.EndTime, timezone).UTC(),
	)

	err = h.slotStore.SaveTimeSlot(timeSlot)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	w.Header().Set("HX-Trigger", "table-update")

	component := templates.SlotFormContent(timeSlot.EventID)
	component.Render(r.Context(), w)
}

// PUT Request on "/slot/{id}"
// Updates time slot in DB to be reserved by user initiating this action
// Redirects to events details endpoint ("/event/{id}")
func (h *SlotHandler) ReserveSlotHandler(w http.ResponseWriter, r *http.Request) {
	timeSlot, err := h.slotStore.GetTimeSlotByVars(r)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	userID := h.authStore.GetUserID(r)

	overlap, err := h.slotStore.CheckUserTimeSlotOverlap(userID, timeSlot.StartTime, timeSlot.EndTime)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	if overlap {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		return
	}

	err = h.slotStore.ReserveTimeSlot(timeSlot, user.ID, fmt.Sprintf("%s %s", user.FirstName, user.LastName))
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/event/%s", timeSlot.EventID))
}

// PUT Request on "/slot/{id}"
// Updates time slot in DB to be free for reservation
// Redirects to events details endpoint ("/event/{id}")
func (h *SlotHandler) FreeSlotHandler(w http.ResponseWriter, r *http.Request) {
	timeSlot, err := h.slotStore.GetTimeSlotByVars(r)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	userID := h.authStore.GetUserID(r)
	h.slotStore.CheckTimeSlotOwnership(w, r, userID, *timeSlot)

	err = h.slotStore.FreeTimeSlot(timeSlot, userID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/event/%s", timeSlot.EventID))
}

// DELETE Request on "/slot/{id}"
// Expected behaviour: delete the time slot from DB
func (h *SlotHandler) DeleteTimeSlotHandler(w http.ResponseWriter, r *http.Request) {
	userID := h.authStore.GetUserID(r)

	timeSlot, err := h.slotStore.GetTimeSlotByVars(r)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	h.slotStore.CheckTimeSlotOwnership(w, r, userID, *timeSlot)

	err = h.slotStore.DeleteTimeSlot(timeSlot.ID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	w.Header().Set("HX-Trigger", "table-update")
}
