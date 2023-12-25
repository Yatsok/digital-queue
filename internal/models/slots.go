package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TimeSlot struct {
	gorm.Model
	ID         uuid.UUID  `json:"id" gorm:"primaryKey"`
	StartTime  time.Time  `json:"startTime" gorm:"type:timestamp without time zone"`
	EndTime    time.Time  `json:"endTime" gorm:"type:timestamp without time zone"`
	UserID     *uuid.UUID `json:"userID" gorm:"type:uuid;null;foreignKey:User"`
	ReservedBy string     `json:"reservedBy"`
	EventID    uuid.UUID  `json:"eventID" gorm:"type:uuid;foreignKey:Event"`
	EventName  string     `json:"eventName"`
}

func NewTimeSlot(userID *uuid.UUID, eventID uuid.UUID, reservedBy, eventName string, startTime, endTime time.Time) *TimeSlot {
	return &TimeSlot{
		ID:         uuid.New(),
		StartTime:  startTime,
		EndTime:    endTime,
		UserID:     userID,
		ReservedBy: reservedBy,
		EventID:    eventID,
		EventName:  eventName,
	}
}
