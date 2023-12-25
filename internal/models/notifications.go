package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	gorm.Model
	ID        uuid.UUID     `json:"id" gorm:"primaryKey"`
	UserID    uuid.UUID     `json:"userID" gorm:"type:uuid;foreignKey:User"`
	EventName string        `json:"eventName" gorm:"foreignKey:Event"`
	TimeLeft  time.Duration `json:"description"`
	IsRead    bool          `json:"isRead"`
}

func NewNotification(userID uuid.UUID, eventName string, timeLeft time.Duration, isRead bool) *Notification {
	return &Notification{
		ID:        uuid.New(),
		UserID:    userID,
		EventName: eventName,
		TimeLeft:  timeLeft,
		IsRead:    isRead,
	}
}
