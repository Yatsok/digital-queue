package controllers

import (
	"net/http"

	"github.com/Yatsok/digital-queue/internal/helper"
	"github.com/Yatsok/digital-queue/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type EventStore interface {
	SaveEvent(event *models.Event) error
	UpdateEvent(eventID uuid.UUID, eventName, eventDescription string) error
	UpdateEventImage(eventID uuid.UUID, imagePath string) error
	DeleteEvent(eventID uuid.UUID) error
	GetEventByID(eventID uuid.UUID) (*models.Event, error)
	GetEventByVars(r *http.Request) (*models.Event, error)
	GetAllEvents() ([]*models.Event, error)
	GetAllEventsByUserID(userID uuid.UUID) ([]*models.Event, error)
	GetPaginatedEvents(offset, limit int) ([]*models.Event, error)
	CountAllEvents() (int, error)
	CalculatePagination(currentPage, totalPages int) []int
	CheckOwnership(userID, ownerID uuid.UUID) bool
	OwnershipRedirect(w http.ResponseWriter, r *http.Request, ownershipStatus bool)
}

type EventService struct {
	db *gorm.DB
}

func NewEventService(db *gorm.DB) *EventService {
	return &EventService{db: db}
}

func (s *EventService) SaveEvent(event *models.Event) error {
	return s.db.Create(event).Error
}

func (s *EventService) UpdateEvent(eventID uuid.UUID, eventName, eventDescription string) error {
	return s.db.Model(&models.Event{}).Where("id = ?", eventID).Updates(map[string]interface{}{
		"name":        eventName,
		"description": eventDescription,
	}).Error
}

func (s *EventService) UpdateEventImage(eventID uuid.UUID, imagePath string) error {
	return s.db.Model(&models.Event{}).Where("id = ?", eventID).Updates(map[string]interface{}{
		"image_path": imagePath,
	}).Error
}

func (s *EventService) DeleteEvent(eventID uuid.UUID) error {
	return s.db.Where("id = ?", eventID).Delete(&models.Event{}).Error
}

func (s *EventService) GetEventByID(eventID uuid.UUID) (*models.Event, error) {
	var event models.Event
	if err := s.db.First(&event, "id = ?", eventID).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *EventService) GetEventByVars(r *http.Request) (*models.Event, error) {
	var event models.Event

	vars := mux.Vars(r)
	eventID := uuid.MustParse(vars["id"])

	if err := s.db.First(&event, "id = ?", eventID).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *EventService) GetAllEvents() ([]*models.Event, error) {
	var events []*models.Event
	if err := s.db.Find(&events).Order("created_at desc").Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (s *EventService) GetAllEventsByUserID(userID uuid.UUID) ([]*models.Event, error) {
	var events []*models.Event
	if err := s.db.Find(&events, "user_id = ?", userID).Order("created_at desc").Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (s *EventService) GetPaginatedEvents(offset, limit int) ([]*models.Event, error) {
	var events []*models.Event
	if err := s.db.Offset(offset).Limit(limit).Find(&events).Order("created_at desc").Error; err != nil {
		return nil, err
	}

	return events, nil
}

func (s *EventService) CountAllEvents() (int, error) {
	var count int64
	if err := s.db.Model(&models.Event{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (s *EventService) CalculatePagination(currentPage, totalPages int) []int {
	return helper.CalculatePagination(currentPage, totalPages)
}

func (s *EventService) CheckOwnership(userID, ownerID uuid.UUID) bool {
	return helper.CheckOwnership(userID, ownerID)
}

func (s *EventService) OwnershipRedirect(w http.ResponseWriter, r *http.Request, ownershipStatus bool) {
	helper.OwnershipRedirect(w, r, ownershipStatus)
}
