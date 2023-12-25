package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	ID          uuid.UUID `json:"id" gorm:"primaryKey"`
	UserID      uuid.UUID `json:"userID" gorm:"type:uuid;foreignKey:ID"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ImagePath   string    `json:"-"`
}

func NewEvent(userID uuid.UUID, name, description string) *Event {
	return &Event{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        name,
		Description: description,
		ImagePath:   "/img/placeholder.png",
	}
}
