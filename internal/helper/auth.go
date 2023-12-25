package helper

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RequestContext struct {
	UserID     string `json:"user_id"`
	AuthStatus bool   `json:"authStatus"`
}

type contextKey string

const (
	AccessTokenLifetime    = time.Minute * 15
	RefreshTokenLifetime   = time.Hour * 24 * 7
	AccessTokenCookieName  = "Access-Token"
	RefreshTokenCookieName = "Refresh-Token"
	requestContextKey      = contextKey("RequestContext")
)

func (c contextKey) String() string {
	return string(c)
}

var (
	JwtSecretKey        = []byte(os.Getenv("JWT_SECRET_KEY"))
	JwtRefreshSecretKey = []byte(os.Getenv("JWT_REFRESH_SECRET_KEY"))
)

func GetJwtSecretKey() ([]byte, error) {
	if len(JwtSecretKey) == 0 {
		return nil, errors.New("JWT secret key is not set")
	}
	return JwtSecretKey, nil
}

func GetJwtRefreshSecretKey() ([]byte, error) {
	if len(JwtRefreshSecretKey) == 0 {
		return nil, errors.New("JWT refresh secret key is not set")
	}
	return JwtRefreshSecretKey, nil
}

func GenerateToken(userID string, secret []byte, lifetime time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(lifetime).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func GenerateTokens(userID string) (string, string, error) {
	newAccessToken, err := GenerateToken(userID, JwtSecretKey, AccessTokenLifetime)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := GenerateToken(userID, JwtRefreshSecretKey, RefreshTokenLifetime)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func VerifyAccessToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return GetJwtSecretKey()
	})

	if err != nil {
		return nil, fmt.Errorf("error verifying access token: %v", err)
	}

	if token.Valid {
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("invalid token claims type")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func VerifyRefreshToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return GetJwtRefreshSecretKey()
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return token.Claims.(jwt.MapClaims), nil
}

func GetRequestContext(r *http.Request) *RequestContext {
	ctxValue := r.Context().Value(requestContextKey)
	if ctxValue == nil {
		return nil
	}
	return ctxValue.(*RequestContext)
}

func SetRequestContext(r *http.Request, authStatus bool, userID string) context.Context {
	requestContext := &RequestContext{
		UserID:     userID,
		AuthStatus: authStatus,
	}
	return context.WithValue(r.Context(), requestContextKey, requestContext)
}

func ExtractClaimFromClaims(key string, claims jwt.MapClaims) string {
	value, _ := claims[key].(string)
	return value
}

func GetAuthStatusContext(w http.ResponseWriter, r *http.Request) bool {
	contextValues := GetRequestContext(r)
	if contextValues == nil {
		http.Error(w, "Unable to retrieve context values", http.StatusInternalServerError)
		return false
	}

	return contextValues.AuthStatus
}

func GetUserIDContext(r *http.Request) string {
	contextValues := GetRequestContext(r)
	if contextValues == nil {
		return ""
	}

	return contextValues.UserID
}

func UpdateCookies(w http.ResponseWriter, newAccessToken, newRefreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:    AccessTokenCookieName,
		Value:   newAccessToken,
		Expires: time.Now().Add(AccessTokenLifetime),
		Path:    "/",
		// Domain:  "127.0.0.1",
	})

	http.SetCookie(w, &http.Cookie{
		Name:    RefreshTokenCookieName,
		Value:   newRefreshToken,
		Expires: time.Now().Add(RefreshTokenLifetime),
		Path:    "/",
		// Domain:  "127.0.0.1",
	})
}

func ClearCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    AccessTokenCookieName,
		Value:   "",
		Expires: time.Now(),
		Path:    "/",
		// Domain:  "127.0.0.1",
		MaxAge: -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:    RefreshTokenCookieName,
		Value:   "",
		Expires: time.Now(),
		Path:    "/",
		// Domain:  "127.0.0.1",
		MaxAge: -1,
	})
}
