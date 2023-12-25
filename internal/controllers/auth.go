package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Yatsok/digital-queue/internal/helper"
	"github.com/Yatsok/digital-queue/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type AuthStore interface {
	CreateUser(user *models.User) error
	AuthenticateUser(email, password string) (*models.User, bool)
	GenerateTokens(userID string) (string, string, error)
	UpdateCookies(w http.ResponseWriter, newAccessToken, newRefreshToken string)
	ClearCookies(w http.ResponseWriter)
	GetAuthStatusContext(w http.ResponseWriter, r *http.Request) bool
	GetUserIDContext(r *http.Request) string
	GetUserID(r *http.Request) uuid.UUID
	SetRequestContext(r *http.Request, authStatus bool, userID string) context.Context
	CheckPermission(w http.ResponseWriter, r *http.Request, authStatus bool)
}

type AuthService struct {
	UserStore UserStore
}

func NewAuthService(UserStore UserStore) *AuthService {
	return &AuthService{UserStore: UserStore}
}

func (s *AuthService) AuthenticateUser(email, password string) (*models.User, bool) {
	user, err := s.UserStore.GetUserByEmail(email)
	if err != nil {
		fmt.Printf("err: %v \n", err)
		return nil, false
	}

	return user, user.ComparePassword(password)
}

func (s *AuthService) UpdateCookies(w http.ResponseWriter, newAccessToken, newRefreshToken string) {
	helper.UpdateCookies(w, newAccessToken, newRefreshToken)
}

func (s *AuthService) ClearCookies(w http.ResponseWriter) {
	helper.ClearCookies(w)
}

func (s *AuthService) CreateUser(user *models.User) error {
	return s.UserStore.CreateUser(user)
}

func (s *AuthService) GenerateTokens(userID string) (string, string, error) {
	return helper.GenerateTokens(userID)
}

func (s *AuthService) GetAuthStatusContext(w http.ResponseWriter, r *http.Request) bool {
	return helper.GetAuthStatusContext(w, r)
}
func (s *AuthService) GetUserIDContext(r *http.Request) string {
	return helper.GetUserIDContext(r)
}
func (s *AuthService) GetUserID(r *http.Request) uuid.UUID {
	userIDContext := s.GetUserIDContext(r)
	userID := uuid.Nil
	if userIDContext != "" {
		userID = uuid.MustParse(userIDContext)
	}

	return userID
}

func (s *AuthService) SetRequestContext(r *http.Request, authStatus bool, userID string) context.Context {
	return helper.SetRequestContext(r, authStatus, userID)
}

func (s *AuthService) CheckPermission(w http.ResponseWriter, r *http.Request, authStatus bool) {
	cookie, err := r.Cookie("OriginalPath")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	checkAuth := func(authNeeded bool, redirectPath string) {
		if authStatus != authNeeded {
			http.Redirect(w, r, redirectPath, http.StatusSeeOther)
		}
	}

	route := mux.CurrentRoute(r)
	if route == nil {
		return
	}

	routeName := route.GetName()
	switch {
	case routeName == "GET_Login" || routeName == "GET_Signup" || routeName == "POST_Login" || routeName == "POST_Signup":
		checkAuth(false, cookie.Value)
	case strings.HasPrefix(r.URL.Path, "/user"):
		checkAuth(true, "/login")
	case routeName == "GET_EventNew" || routeName == "POST_EventNew":
		checkAuth(true, "/login")
	case strings.HasPrefix(r.URL.Path, "/event"):
		if routeName == "GET_Event" || routeName == "GET_Events" || routeName == "GET_Events_Paginated" {
			return
		}
		checkAuth(true, "/login")
	case strings.HasPrefix(r.URL.Path, "/slot"):
		checkAuth(true, "/login")
	case strings.HasPrefix(r.URL.Path, "/upload"):
		checkAuth(true, "/login")
	}
}
