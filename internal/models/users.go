package models

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID           uuid.UUID `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"unique"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	Password     string    `json:"-"`
	Country      string    `json:"country"`
	Timezone     string    `json:"timezone"`
	IsSubscribed bool      `json:"isSubscribed"`
	ImagePath    string    `json:"-"`
}

func NewUser(email, firstName, lastName, password, timezone string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           uuid.New(),
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		Password:     string(hashedPassword),
		Country:      "N/A",
		Timezone:     timezone,
		IsSubscribed: true,
		ImagePath:    "/img/user.jpg",
	}, nil
}

func (u *User) ComparePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) HashPassword(password string) []byte {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing new password")
	}
	return hashedPassword
}
