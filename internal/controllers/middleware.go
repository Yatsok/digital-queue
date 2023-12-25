package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Yatsok/digital-queue/internal/helper"
)

type AuthMiddlewareStore interface {
	CheckAuthentication(r *http.Request) (bool, string)
	RefreshTokens(w http.ResponseWriter, r *http.Request) (bool, string)
	SetRequestContext(r *http.Request, authStatus bool, userID string) context.Context
	UpdateCookies(w http.ResponseWriter, newAccessToken, newRefreshToken string)
	ClearCookies(w http.ResponseWriter)
}

type AuthMiddlewareService struct {
	AuthStore AuthStore
}

func NewAuthMiddlewareService(AuthStore AuthStore) *AuthMiddlewareService {
	return &AuthMiddlewareService{AuthStore: AuthStore}
}

type contextKey string

func (c contextKey) String() string {
	return "auth context key: " + string(c)
}

const (
	AccessTokenCookieName             = "Access-Token"
	RefreshTokenCookieName            = "Refresh-Token"
	AuthStatusKey          contextKey = "auth_status"
	UserIDKey              contextKey = "user_id"
)

func (s *AuthMiddlewareService) CheckAuthentication(r *http.Request) (bool, string) {
	authToken, err := r.Cookie(AccessTokenCookieName)
	if err != nil || authToken == nil {
		return false, ""
	}

	claims, err := helper.VerifyAccessToken(authToken.Value)
	if err != nil {
		return false, ""
	}

	userID := helper.ExtractClaimFromClaims("user_id", claims)

	return true, userID
}

func (s *AuthMiddlewareService) RefreshTokens(w http.ResponseWriter, r *http.Request) (bool, string) {
	refreshToken, err := r.Cookie(RefreshTokenCookieName)

	if err != nil || refreshToken == nil {
		return false, ""
	}

	claims, err := helper.VerifyRefreshToken(refreshToken.Value)
	if err != nil {
		fmt.Printf("Error verifying refresh token: %v\n", err)
		return false, ""
	}

	userID := helper.ExtractClaimFromClaims("user_id", claims)
	if userID == "" {
		fmt.Println("User ID claim is empty")
		return false, ""
	}

	newAccessToken, newRefreshToken, err := s.AuthStore.GenerateTokens(userID)
	if err != nil {
		return false, ""
	}

	s.ClearCookies(w)
	s.UpdateCookies(w, newAccessToken, newRefreshToken)

	return true, userID
}

func (s *AuthMiddlewareService) SetRequestContext(r *http.Request, authStatus bool, userID string) context.Context {
	return helper.SetRequestContext(r, authStatus, userID)
}

func (s *AuthMiddlewareService) UpdateCookies(w http.ResponseWriter, newAccessToken, newRefreshToken string) {
	helper.UpdateCookies(w, newAccessToken, newRefreshToken)
}

func (s *AuthMiddlewareService) ClearCookies(w http.ResponseWriter) {
	helper.ClearCookies(w)
}
