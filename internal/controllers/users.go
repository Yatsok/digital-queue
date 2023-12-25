package controllers

import (
	"fmt"
	"net/http"

	"github.com/Yatsok/digital-queue/internal/helper"
	"github.com/Yatsok/digital-queue/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type UserStore interface {
	CreateUser(user *models.User) error
	GetUserByID(id uuid.UUID) (*models.User, error)
	GetUserByVars(r *http.Request) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(user *models.User) error
	UpdateUserImage(userID uuid.UUID, imagePath string) error
	DeleteUser(id uuid.UUID) error
	GetAuthStatusContext(w http.ResponseWriter, r *http.Request) bool
	GetUserIDContext(r *http.Request) string
	GetUserTimezoneByID(id uuid.UUID) (string, error)
	GetUserTimezone(id uuid.UUID) string
}

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

func (s *UserService) CreateUser(user *models.User) error {
	return s.db.Create(user).Error
}

func (s *UserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.First(&user, "id = ?", id).Error
	return &user, err
}

func (s *UserService) GetUserByVars(r *http.Request) (*models.User, error) {
	var user models.User

	vars := mux.Vars(r)
	userID := uuid.MustParse(vars["id"])

	err := s.db.First(&user, "id = ?", userID).Error
	return &user, err
}

func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := s.db.First(&user, "email = ?", email).Error
	return &user, err
}

func (s *UserService) UpdateUser(user *models.User) error {
	return s.db.Save(user).Error
}

func (s *UserService) UpdateUserImage(userID uuid.UUID, imagePath string) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"image_path": imagePath,
	}).Error
}

func (s *UserService) DeleteUser(id uuid.UUID) error {
	return s.db.Delete(&models.User{}, id).Error
}

func (s *UserService) GetAuthStatusContext(w http.ResponseWriter, r *http.Request) bool {
	return helper.GetAuthStatusContext(w, r)
}

func (s *UserService) GetUserIDContext(r *http.Request) string {
	return helper.GetUserIDContext(r)
}

func (s *UserService) GetUserTimezoneByID(id uuid.UUID) (string, error) {
	var user models.User
	err := s.db.Select("timezone").First(&user, "id = ?", id).Error
	return user.Timezone, err
}

func (s *UserService) GetUserTimezone(userID uuid.UUID) string {
	timezone, err := s.GetUserTimezoneByID(userID)
	if err != nil {
		fmt.Println("Error fetching timezone")
		timezone = "UTC"
	}
	return timezone
}
