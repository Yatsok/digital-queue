package handlers

import (
	"net/http"
	"strings"

	"github.com/Yatsok/digital-queue/internal/controllers"
	"github.com/Yatsok/digital-queue/internal/helper"
)

type AuthMiddlewareHandler struct {
	AuthMiddlewareStore controllers.AuthMiddlewareStore
	AuthStore           controllers.AuthStore
}

func NewAuthMiddlewareHandler(AuthMiddlewareStore controllers.AuthMiddlewareStore, AuthStore controllers.AuthStore) *AuthMiddlewareHandler {
	return &AuthMiddlewareHandler{
		AuthMiddlewareStore: AuthMiddlewareStore,
		AuthStore:           AuthStore,
	}
}

func (h *AuthMiddlewareHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		excludedPaths := []string{"/css", "/img"}
		for _, path := range excludedPaths {
			if strings.HasPrefix(r.URL.Path, path) {
				next.ServeHTTP(w, r)
				return
			}
		}

		cookie, err := r.Cookie("OriginalPath")
		if err != nil && cookie == nil {
			http.SetCookie(w, &http.Cookie{
				Name:  "OriginalPath",
				Value: "/",
				Path:  "/",
			})
		}

		authStatus, userID := h.AuthMiddlewareStore.CheckAuthentication(r)
		if !authStatus {
			authStatus, userID = h.AuthMiddlewareStore.RefreshTokens(w, r)
		}

		h.AuthStore.CheckPermission(w, r, authStatus)

		http.SetCookie(w, &http.Cookie{
			Name:  "OriginalPath",
			Value: r.URL.Path,
			Path:  "/",
		})

		ctx := helper.SetRequestContext(r, authStatus, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
