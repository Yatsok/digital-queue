package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Yatsok/digital-queue/internal/controllers"
	"github.com/Yatsok/digital-queue/internal/models"
	"github.com/Yatsok/digital-queue/templates"
)

type AuthHandler struct {
	AuthStore controllers.AuthStore
	CustomErrorHandler
}

func NewAuthHandler(AuthStore controllers.AuthStore) *AuthHandler {
	return &AuthHandler{AuthStore: AuthStore}
}

func (h *AuthHandler) SigninHandler(w http.ResponseWriter, r *http.Request) {

	component := templates.Base(templates.SignIn())
	component.Render(r.Context(), w)
}

func (h *AuthHandler) SignupHandler(w http.ResponseWriter, r *http.Request) {

	component := templates.Base(templates.SignUp())
	component.Render(r.Context(), w)
}

func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Password != req.ConfirmPassword {
		http.Error(w, "Password and Confirm Password do not match", http.StatusBadRequest)
		return
	}

	user, err := models.NewUser(req.Email, req.FirstName, req.LastName, req.Password, req.Timezone)
	if err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	if err := h.AuthStore.CreateUser(user); err != nil {
		h.CustomErrorHandler.InternalServerErrorHandler(w, r)
		return
	}

	w.Header().Set("HX-Redirect", "/login")
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	user, authenticated := h.AuthStore.AuthenticateUser(req.Email, req.Password)
	if authenticated {
		token, refreshToken, err := h.AuthStore.GenerateTokens(user.ID.String())
		if err != nil {
			h.CustomErrorHandler.InternalServerErrorHandler(w, r)
			return
		}

		h.AuthStore.UpdateCookies(w, token, refreshToken)

		w.Header().Set("HX-Redirect", "/")
		return
	}

	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}

func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	h.AuthStore.ClearCookies(w)

	w.Header().Set("HX-Redirect", "/login")
}
