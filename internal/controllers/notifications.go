package controllers

import (
	"net/http"

	"github.com/Yatsok/digital-queue/internal/helper"
	"github.com/Yatsok/digital-queue/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type NotificationStore interface {
	CreateNotification(nt *models.Notification) error
	GetNotificationByID(id uuid.UUID) (*models.Notification, error)
	GetNotificationsByUserID(userID uuid.UUID) ([]*models.Notification, error)
	GetNotificationByVars(r *http.Request) (*models.Notification, error)
	UpdateNotification(nt *models.Notification) error
	DeleteNotification(id uuid.UUID) error
	CheckOwnership(userID, ownerID uuid.UUID) bool
	OwnershipRedirect(w http.ResponseWriter, r *http.Request, ownershipStatus bool)
}

type NotificationService struct {
	db *gorm.DB
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{
		db: db,
	}
}

func (s *NotificationService) CreateNotification(nt *models.Notification) error {
	return s.db.Create(nt).Error
}

func (s *NotificationService) GetNotificationByID(id uuid.UUID) (*models.Notification, error) {
	notification := new(models.Notification)
	err := s.db.First(notification, id).Error
	if err != nil {
		return nil, err
	}
	return notification, nil
}

func (s *NotificationService) GetNotificationsByUserID(userID uuid.UUID) ([]*models.Notification, error) {
	notifications := make([]*models.Notification, 0)
	err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (s *NotificationService) GetNotificationByVars(r *http.Request) (*models.Notification, error) {
	var notification models.Notification

	vars := mux.Vars(r)
	ntIDContext := vars["id"]

	ntID := uuid.Nil
	if ntIDContext != "" {
		ntID = uuid.MustParse(ntIDContext)
	}

	if err := s.db.First(&notification, "id = ?", ntID).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

func (s *NotificationService) UpdateNotification(nt *models.Notification) error {
	return s.db.Save(nt).Error
}

func (s *NotificationService) DeleteNotification(id uuid.UUID) error {
	return s.db.Delete(&models.Notification{}, id).Error
}

func (s *NotificationService) CheckOwnership(userID, ownerID uuid.UUID) bool {
	return helper.CheckOwnership(userID, ownerID)
}

func (s *NotificationService) OwnershipRedirect(w http.ResponseWriter, r *http.Request, ownershipStatus bool) {
	helper.OwnershipRedirect(w, r, ownershipStatus)
}
