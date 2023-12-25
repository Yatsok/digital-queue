package models

import "github.com/google/uuid"

type AuthRequest struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Timezone        string `json:"timezone"`
}

type EventRequestGET struct {
	AuthStatus      bool
	OwnershipStatus bool
	UserID          uuid.UUID
	Event           Event
	Events          []*Event
	Page            int
	TotalPages      int
	Pagination      []int
	TimeSlots       map[string][]TimeSlot
	Timezone        string
	ImagePath       string
}

type EventRequestPOST struct {
	ID          string `json:"event_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SlotRequestPOST struct {
	EventID   string `json:"event_id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type SlotRequestGET struct {
	AuthStatus      bool
	OwnershipStatus bool
	UserID          uuid.UUID
	Event           Event
	TimeSlots       []TimeSlot
	Timezone        string
}

type UserRequestGET struct {
	AuthStatus   bool
	User         User
	Events       []*Event
	TimeSlotsMap map[string][]TimeSlot
	TimeSlots    []TimeSlot
}

type UserRequestPUT struct {
	ID              string `json:"event_id"`
	Email           string `json:"email"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	Country         string `json:"country"`
	Timezone        string `json:"timezone"`
	IsSubscribed    string `json:"isSubscribed"`
	Toggle          string `json:"toggle"`
	OldPassword     string `json:"oldPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}
