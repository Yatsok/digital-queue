package handlers

import (
	"net/http"

	"github.com/Yatsok/digital-queue/templates"
)

type CustomErrorHandler struct {
}

func NewCustomErrorHandler() *CustomErrorHandler {
	return &CustomErrorHandler{}
}

func (h *CustomErrorHandler) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	component := templates.Base(templates.NotFoundPage())
	component.Render(r.Context(), w)
}

func (h *CustomErrorHandler) InternalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)

	component := templates.Base(templates.InternalServerErrorPage())
	component.Render(r.Context(), w)
}
