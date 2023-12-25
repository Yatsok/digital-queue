package controllers

import (
	"net/http"
	"sort"
	"time"

	"github.com/Yatsok/digital-queue/internal/helper"
	"github.com/Yatsok/digital-queue/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type SlotStore interface {
	SaveTimeSlot(timeSlot *models.TimeSlot) error
	DeleteTimeSlot(timeSlotID uuid.UUID) error
	ReserveTimeSlot(timeSlot *models.TimeSlot, userID uuid.UUID, reservedBy string) error
	UpdateTimeSlot(timeSlot *models.TimeSlot, userID uuid.UUID, reservedBy string) error
	FreeTimeSlot(timeSlot *models.TimeSlot, nilUserID uuid.UUID) error
	GetTimeSlotsByEventID(eventID uuid.UUID) ([]models.TimeSlot, error)
	GetAllTimeSlotsReservedByUserID(userID uuid.UUID) ([]models.TimeSlot, error)
	GetTimeSlotByID(timeSlotID uuid.UUID) (*models.TimeSlot, error)
	GetTimeSlotByVars(r *http.Request) (*models.TimeSlot, error)
	GetAllTimeSlots() ([]models.TimeSlot, error)
	GetFilteredTimeSlots(allTimeSlots map[string][]models.TimeSlot, userID uuid.UUID) map[string][]models.TimeSlot
	GetUpcomingTimeSlots(currentTime time.Time, threshold time.Duration) ([]models.TimeSlot, error)
	GetExpiredTimeSlots(currentTime time.Time) ([]models.TimeSlot, error)
	CheckUserTimeSlotOverlap(userID uuid.UUID, startTime, endTime time.Time) (bool, error)
	CheckTimeSlotOwnership(w http.ResponseWriter, r *http.Request, userID uuid.UUID, timeSlot models.TimeSlot)
	CheckOwnership(userID, ownerID uuid.UUID) bool
	OwnershipRedirect(w http.ResponseWriter, r *http.Request, ownershipStatus bool)
}

type SlotService struct {
	db *gorm.DB
}

func NewSlotService(db *gorm.DB) *SlotService {
	return &SlotService{db: db}
}

func (s *SlotService) SaveTimeSlot(timeSlot *models.TimeSlot) error {
	return s.db.Create(timeSlot).Error
}

func (s *SlotService) DeleteTimeSlot(timeSlotID uuid.UUID) error {
	return s.db.Where("id = ?", timeSlotID).Delete(&models.TimeSlot{}).Error
}

func (s *SlotService) ReserveTimeSlot(timeSlot *models.TimeSlot, userID uuid.UUID, reservedBy string) error {
	return s.db.Model(&models.TimeSlot{}).
		Where("id = ? AND user_id IS NULL", timeSlot.ID).
		Updates(map[string]interface{}{"user_id": userID, "reserved_by": reservedBy}).
		Error
}

func (s *SlotService) UpdateTimeSlot(timeSlot *models.TimeSlot, userID uuid.UUID, reservedBy string) error {
	return s.db.Model(&models.TimeSlot{}).
		Where("id = ? AND user_id = ?", timeSlot.ID, userID).
		Updates(map[string]interface{}{"user_id": userID, "reserved_by": reservedBy}).
		Error
}

func (s *SlotService) FreeTimeSlot(timeSlot *models.TimeSlot, userID uuid.UUID) error {
	return s.db.Model(&models.TimeSlot{}).
		Where("id = ? AND user_id = ?", timeSlot.ID, userID).
		Updates(map[string]interface{}{"user_id": nil, "reserved_by": ""}).
		Error
}

func (s *SlotService) GetTimeSlotsByEventID(eventID uuid.UUID) ([]models.TimeSlot, error) {
	var timeSlots []models.TimeSlot
	if err := s.db.Find(&timeSlots, "event_id = ?", eventID).Order("start_time asc").Error; err != nil {
		return nil, err
	}
	return timeSlots, nil
}

func (s *SlotService) GetAllTimeSlotsReservedByUserID(userID uuid.UUID) ([]models.TimeSlot, error) {
	var timeSlots []models.TimeSlot
	if err := s.db.Find(&timeSlots, "user_id = ?", userID).Order("start_time asc").Error; err != nil {
		return nil, err
	}
	return timeSlots, nil
}

func (s *SlotService) GetAllTimeSlots() ([]models.TimeSlot, error) {
	var timeSlots []models.TimeSlot
	if err := s.db.Find(&timeSlots).Order("start_time asc").Error; err != nil {
		return nil, err
	}
	return timeSlots, nil
}

func (s *SlotService) GetFilteredTimeSlots(allTimeSlots map[string][]models.TimeSlot, userID uuid.UUID) map[string][]models.TimeSlot {
	filteredTimeSlots := make(map[string][]models.TimeSlot)

	for eventID, timeSlots := range allTimeSlots {
		var filteredSlots []models.TimeSlot

		for _, slot := range timeSlots {
			if slot.UserID == nil || (userID != uuid.Nil && slot.UserID.String() == userID.String()) {
				filteredSlots = append(filteredSlots, slot)
			}
		}

		sort.Slice(filteredSlots, func(i, j int) bool {
			return filteredSlots[i].StartTime.Before(filteredSlots[j].StartTime)
		})

		filteredTimeSlots[eventID] = filteredSlots
	}

	return filteredTimeSlots
}

func (s *SlotService) GetTimeSlotByID(timeSlotID uuid.UUID) (*models.TimeSlot, error) {
	var timeSlot models.TimeSlot
	if err := s.db.First(&timeSlot, "id = ?", timeSlotID).Error; err != nil {
		return nil, err
	}
	return &timeSlot, nil
}

func (s *SlotService) GetTimeSlotByVars(r *http.Request) (*models.TimeSlot, error) {
	vars := mux.Vars(r)
	timeSlotID := uuid.MustParse(vars["id"])

	var timeSlot models.TimeSlot
	if err := s.db.First(&timeSlot, "id = ?", timeSlotID).Error; err != nil {
		return nil, err
	}
	return &timeSlot, nil
}

func (s *SlotService) GetUpcomingTimeSlots(currentTime time.Time, threshold time.Duration) ([]models.TimeSlot, error) {
	var timeSlots []models.TimeSlot
	if err := s.db.Find(&timeSlots, "start_time >= ? AND start_time <= ?", currentTime, currentTime.Add(threshold)).Order("start_time asc").Error; err != nil {
		return nil, err
	}
	return timeSlots, nil
}

func (s *SlotService) GetExpiredTimeSlots(currentTime time.Time) ([]models.TimeSlot, error) {
	var timeSlots []models.TimeSlot
	if err := s.db.Find(&timeSlots, "end_time <= ?", currentTime).Order("start_time asc").Error; err != nil {
		return nil, err
	}
	return timeSlots, nil
}

func (s *SlotService) CheckUserTimeSlotOverlap(userID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	userTimeSlots, err := s.GetAllTimeSlotsReservedByUserID(userID)
	if err != nil {
		return false, err
	}

	for _, ts := range userTimeSlots {
		if (startTime.Before(ts.EndTime) || startTime.Equal(ts.EndTime)) &&
			(endTime.After(ts.StartTime) || endTime.Equal(ts.StartTime)) {
			return true, nil
		}
	}

	return false, nil
}

func (s *SlotService) CheckOwnership(userID, ownerID uuid.UUID) bool {
	return helper.CheckOwnership(userID, ownerID)
}

func (s *SlotService) OwnershipRedirect(w http.ResponseWriter, r *http.Request, ownershipStatus bool) {
	helper.OwnershipRedirect(w, r, ownershipStatus)
}

func (s *SlotService) CheckTimeSlotOwnership(w http.ResponseWriter, r *http.Request, userID uuid.UUID, timeSlot models.TimeSlot) {
	event, err := s.GetEventByID(timeSlot.EventID)
	if err != nil {
		return
	}

	ownershipStatus := s.CheckOwnership(userID, event.UserID)
	s.OwnershipRedirect(w, r, ownershipStatus)
}

func (s *SlotService) GetEventByID(eventID uuid.UUID) (*models.Event, error) {
	var event models.Event
	if err := s.db.First(&event, "id = ?", eventID).Error; err != nil {
		return nil, err
	}
	return &event, nil
}
