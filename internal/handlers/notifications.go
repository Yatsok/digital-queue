package handlers

import (
	"net/http"

	"github.com/Yatsok/digital-queue/internal/controllers"
	"github.com/Yatsok/digital-queue/templates"
)

type NotificationHandler struct {
	notificationStore controllers.NotificationStore
	authStore         controllers.AuthStore
	CustomErrorHandler
}

func NewNotificationHandler(notificationStore controllers.NotificationStore, authStore controllers.AuthStore) *NotificationHandler {
	return &NotificationHandler{
		notificationStore: notificationStore,
		authStore:         authStore,
	}
}

// GET Request on "/notifications"
// Getes all notifications from DB
// Render templ component NotificationsList
func (h *NotificationHandler) NotificationListHandler(w http.ResponseWriter, r *http.Request) {
	userID := h.authStore.GetUserID(r)

	userNotifications, err := h.notificationStore.GetNotificationsByUserID(userID)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	component := templates.Base(templates.NotificationsList(userNotifications))
	component.Render(r.Context(), w)
}

func (h *NotificationHandler) NotificationHideHandler(w http.ResponseWriter, r *http.Request) {
	userID := h.authStore.GetUserID(r)

	notification, err := h.notificationStore.GetNotificationByVars(r)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	ownershipStatus := h.notificationStore.CheckOwnership(userID, notification.UserID)
	h.notificationStore.OwnershipRedirect(w, r, ownershipStatus)

	notification.IsRead = true

	h.notificationStore.UpdateNotification(notification)
}
