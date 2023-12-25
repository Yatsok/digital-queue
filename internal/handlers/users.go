package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Yatsok/digital-queue/internal/controllers"
	"github.com/Yatsok/digital-queue/internal/models"
	"github.com/Yatsok/digital-queue/templates"
)

type UserHandler struct {
	userStore  controllers.UserStore
	authStore  controllers.AuthStore
	eventStore controllers.EventStore
	slotStore  controllers.SlotStore
	CustomErrorHandler
}

func NewUserHandler(userStore controllers.UserStore, authStore controllers.AuthStore, eventStore controllers.EventStore, slotStore controllers.SlotStore) *UserHandler {
	return &UserHandler{
		userStore:  userStore,
		authStore:  authStore,
		eventStore: eventStore,
		slotStore:  slotStore,
	}
}

// GET Request on "/user"
// Render templ component UserProfile
func (h *UserHandler) UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := h.authStore.GetUserID(r)

	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	req := &models.UserRequestGET{
		AuthStatus: h.authStore.GetAuthStatusContext(w, r),
		User:       *user,
	}

	component := templates.Base(templates.UserProfile(req))
	component.Render(r.Context(), w)
}

// DELETE Request on "/user/{id}"
func (h *UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := h.authStore.GetUserID(r)

	h.userStore.DeleteUser(userID)

	events, err := h.eventStore.GetAllEventsByUserID(userID)
	if err != nil {
		return
	}

	for _, event := range events {
		h.eventStore.DeleteEvent(event.ID)
	}

	userSlots, err := h.slotStore.GetAllTimeSlotsReservedByUserID(userID)
	if err != nil {
		return
	}

	for _, userSlot := range userSlots {
		h.slotStore.FreeTimeSlot(&userSlot, userID)
	}

	h.authStore.ClearCookies(w)
	w.Header().Set("HX-Redirect", "/login")
}

// GET Request on "/user/events"
// Fetches all events created by user from DB
// Render templ component UserEvents
func (h *UserHandler) UserEventsHandler(w http.ResponseWriter, r *http.Request) {
	userID := h.authStore.GetUserID(r)

	events, err := h.eventStore.GetAllEventsByUserID(userID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	allTimeSlots := make(map[string][]models.TimeSlot)
	for _, event := range events {
		timeSlots, err := h.slotStore.GetTimeSlotsByEventID(event.ID)
		if err != nil {
			h.CustomErrorHandler.InternalServerErrorHandler(w, r)
			return
		}
		allTimeSlots[event.ID.String()] = timeSlots
	}

	req := &models.UserRequestGET{
		AuthStatus:   h.authStore.GetAuthStatusContext(w, r),
		Events:       events,
		TimeSlotsMap: allTimeSlots,
	}

	component := templates.Base(templates.UserEvents(req))
	component.Render(r.Context(), w)
}

// GET Request on "/user/slots"
// Fetches all time slots reserved by user from DB
// Render templ component UserEvents
func (h *UserHandler) UserSlotsHandler(w http.ResponseWriter, r *http.Request) {
	userID := h.authStore.GetUserID(r)
	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		return
	}

	usertimeSlots, err := h.slotStore.GetAllTimeSlotsReservedByUserID(userID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	req := &models.UserRequestGET{
		AuthStatus: h.authStore.GetAuthStatusContext(w, r),
		TimeSlots:  usertimeSlots,
		User:       *user,
	}

	component := templates.Base(templates.UserSlots(req))
	component.Render(r.Context(), w)
}

// GET Request on "/user/settings"
// Render templ component UserSettings
func (h *UserHandler) UserSettingsHandler(w http.ResponseWriter, r *http.Request) {
	userID := h.authStore.GetUserID(r)

	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	req := &models.UserRequestGET{
		AuthStatus: h.authStore.GetAuthStatusContext(w, r),
		User:       *user,
	}

	component := templates.Base(templates.UserSettings(req))
	component.Render(r.Context(), w)
}

// GET Request on "/user/account"
// Render templ component SettingsAccount
func (h *UserHandler) UserAccountSettingsHandler(w http.ResponseWriter, r *http.Request) {
	userID := h.authStore.GetUserID(r)

	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	req := &models.UserRequestGET{
		User: *user,
	}

	component := templates.Base(templates.SettingsAccount(req))
	component.Render(r.Context(), w)
}

// PUT Request on "/user/account"
func (h *UserHandler) ChangeAccountSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var req models.UserRequestPUT
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	userID := h.authStore.GetUserID(r)

	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Email = req.Email
	user.Country = req.Country
	user.Timezone = req.Timezone

	h.userStore.UpdateUser(user)

	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)

	userSlots, err := h.slotStore.GetAllTimeSlotsReservedByUserID(userID)
	if err != nil {
		return
	}

	for _, userSlot := range userSlots {
		err := h.slotStore.UpdateTimeSlot(&userSlot, userID, fullName)
		if err != nil {
			fmt.Printf("err: %v \n", err)
			return
		}
	}

	w.Header().Set("HX-Redirect", "/user")
}

// GET Request on "/user/security"
// Render templ component SettingsSecurity
func (h *UserHandler) UserSecuritySettingsHandler(w http.ResponseWriter, r *http.Request) {
	component := templates.Base(templates.SettingsSecurity())
	component.Render(r.Context(), w)
}

// PUT Request on "/user/security"
func (h *UserHandler) ChangeSecuritySettingsHandler(w http.ResponseWriter, r *http.Request) {
	var req models.UserRequestPUT
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	userID := h.authStore.GetUserID(r)

	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	if user.ComparePassword(req.OldPassword) {
		fmt.Println("Old password verified")
		if req.NewPassword == req.ConfirmPassword {
			fmt.Println("New password confirmed")
			user.Password = string(user.HashPassword(req.NewPassword))
			if err := h.userStore.UpdateUser(user); err != nil {
				fmt.Println("Error updating user:", err)
				h.CustomErrorHandler.InternalServerErrorHandler(w, r)
				return
			}
		} else {
			fmt.Println("New password confirmation failed")
			return
		}
	} else {
		fmt.Println("Old password verification failed")
		return
	}

	w.Header().Set("HX-Redirect", "/user")
}

// GET Request on "/user/appearance"
// Render templ component SettingsAppearance
func (h *UserHandler) UserAppearanceSettingsHandler(w http.ResponseWriter, r *http.Request) {
	userID := h.authStore.GetUserID(r)
	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		return
	}

	req := &models.UserRequestGET{
		User: *user,
	}

	component := templates.Base(templates.SettingsAppearance(req))
	component.Render(r.Context(), w)
}

// GET Request on "/user/notifications"
// Render templ component SettingsNotifications
func (h *UserHandler) UserNotificationsSettingsHandler(w http.ResponseWriter, r *http.Request) {
	userID := h.authStore.GetUserID(r)

	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	req := &models.UserRequestGET{
		User: *user,
	}

	component := templates.Base(templates.SettingsNotifications(req))
	component.Render(r.Context(), w)
}

// PUT Request on "/user/notifications"
func (h *UserHandler) ChangeNotificationsSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var req models.UserRequestPUT
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	userID := h.authStore.GetUserID(r)

	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	user.IsSubscribed, err = strconv.ParseBool(req.IsSubscribed)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	h.userStore.UpdateUser(user)

	w.Header().Set("HX-Redirect", "/user")
}
