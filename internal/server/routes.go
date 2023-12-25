package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Yatsok/digital-queue/internal/helper"
	"github.com/Yatsok/digital-queue/templates"
	"github.com/google/uuid"
)

func (s *Server) RegisterRoutes() http.Handler {
	// Middleware
	s.Router.Use(s.AuthMiddlewareHandler.AuthMiddleware)

	// Home Page
	home := homeHandler{}
	s.Router.HandleFunc("/", home.ServeHTTP)

	// Static Files
	jsHandler := http.StripPrefix("/js/", http.FileServer(http.Dir("./assets/js/")))
	cssHandler := http.StripPrefix("/css/", http.FileServer(http.Dir("./assets/css/")))
	imgHandler := http.StripPrefix("/img/", http.FileServer(http.Dir("./assets/img/")))
	s.Router.PathPrefix("/js/").Handler(jsHandler)
	s.Router.PathPrefix("/css/").Handler(cssHandler)
	s.Router.PathPrefix("/img/").Handler(imgHandler)

	// Auth Handlers
	s.Router.HandleFunc("/login", s.AuthHandler.SigninHandler).Methods("GET").Name("GET_Login")
	s.Router.HandleFunc("/signup", s.AuthHandler.SignupHandler).Methods("GET").Name("GET_Signup")
	s.Router.HandleFunc("/signup", s.AuthHandler.RegisterHandler).Methods("POST").Name("POST_Signup")
	s.Router.HandleFunc("/login", s.AuthHandler.LoginHandler).Methods("POST").Name("POST_Login")
	s.Router.HandleFunc("/logout", s.AuthHandler.LogoutHandler).Methods("POST")

	// Event Handlers
	s.Router.HandleFunc("/event/new", s.EventHandler.CreateEventFormHandler).Methods("GET").Name("GET_EventNew")
	s.Router.HandleFunc("/event/{id}/upload", s.EventHandler.UploadImageFormHandler).Methods("GET").Name("GET_EventUpload")
	s.Router.HandleFunc("/event", s.EventHandler.CreateEventHandler).Methods("POST").Name("POST_Event")
	s.Router.HandleFunc("/event/{id}/edit", s.EventHandler.EditEventFormHandler).Methods("GET").Name("GET_EventEdit")
	s.Router.HandleFunc("/event", s.EventHandler.EditEventHandler).Methods("PUT").Name("PUT_Event")
	s.Router.HandleFunc("/event/{id}", s.EventHandler.EventDetailsHandler).Methods("GET").Name("GET_Event")
	s.Router.HandleFunc("/events", s.EventHandler.EventsListHandler).Methods("GET").Name("GET_Events")
	s.Router.HandleFunc("/events/{page}", s.EventHandler.EventsListHandler).Methods("GET").Name("GET_Events_Paginated")
	s.Router.HandleFunc("/event/{id}", s.EventHandler.DeleteEventHandler).Methods("DELETE").Name("DELETE_Event")

	// Slot Handlers
	s.Router.HandleFunc("/slot/{id}/reserve", s.SlotHandler.ReserveSlotHandler).Methods("PUT")
	s.Router.HandleFunc("/slot/{id}/cancel", s.SlotHandler.FreeSlotHandler).Methods("PUT")
	s.Router.HandleFunc("/event/{id}/add", s.SlotHandler.CreateTimeSlotsFormHandler).Name("СreateSlots").Methods("GET")
	s.Router.HandleFunc("/event/{id}/slots", s.EventHandler.EventSlotsHandler).Name("SlotsTable").Methods("GET")
	s.Router.HandleFunc("/event/add", s.SlotHandler.CreateTimeSlotHandler).Methods("POST")
	s.Router.HandleFunc("/slot/{id}", s.SlotHandler.DeleteTimeSlotHandler).Methods("DELETE")

	// User Handlers
	s.Router.HandleFunc("/user", s.UserHandler.UserProfileHandler).Methods("GET").Name("GET_User")
	s.Router.HandleFunc("/user/{id}", s.UserHandler.DeleteUserHandler).Methods("DELETE").Name("DELETE_User")
	s.Router.HandleFunc("/user/events", s.UserHandler.UserEventsHandler).Methods("GET").Name("GET_UserEvents")
	s.Router.HandleFunc("/user/slots", s.UserHandler.UserSlotsHandler).Methods("GET").Name("GET_UserSlots")
	s.Router.HandleFunc("/user/settings", s.UserHandler.UserSettingsHandler).Methods("GET").Name("GET_UserSettings")
	s.Router.HandleFunc("/user/account", s.UserHandler.UserAccountSettingsHandler).Methods("GET").Name("GET_UserAccountSettings")
	s.Router.HandleFunc("/user/account", s.UserHandler.ChangeAccountSettingsHandler).Methods("PUT").Name("PUT_UserAccountSettings")
	s.Router.HandleFunc("/user/security", s.UserHandler.UserSecuritySettingsHandler).Methods("GET").Name("GET_UserSecuritySettings")
	s.Router.HandleFunc("/user/security", s.UserHandler.ChangeSecuritySettingsHandler).Methods("PUT").Name("PUT_UserSecuritySettings")
	s.Router.HandleFunc("/user/appearance", s.UserHandler.UserAppearanceSettingsHandler).Methods("GET").Name("GET_UserAppearanceSettings")
	s.Router.HandleFunc("/user/notifications", s.UserHandler.UserNotificationsSettingsHandler).Methods("GET").Name("GET_UserNotificationsSettings")
	s.Router.HandleFunc("/user/notifications", s.UserHandler.ChangeNotificationsSettingsHandler).Methods("PUT").Name("PUT_UserNotificationsSettings")

	// Notification Handlers
	s.Router.HandleFunc("/ws", s.handleWebSocket)
	s.Router.HandleFunc("/notifications", s.NotificationHandler.NotificationListHandler).Methods("GET")
	s.Router.HandleFunc("/notifications/{id}", s.NotificationHandler.NotificationHideHandler).Methods("PUT")

	// DB Handlers
	s.Router.HandleFunc("/health", s.healthHandler)
	s.Router.HandleFunc("/upload", s.UploadHandler).Methods("POST")

	// 404 Page
	s.Router.NotFoundHandler = http.HandlerFunc(s.CustomErrorHandler.NotFoundHandler)

	return s.Router
}

type homeHandler struct{}

func (h *homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authStatus := helper.GetAuthStatusContext(w, r)

	component := templates.Base(templates.IndexPage(authStatus))
	component.Render(r.Context(), w)
}

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	authStatus := helper.GetAuthStatusContext(w, r)

	component := templates.Base(templates.IndexPage(authStatus))
	component.Render(r.Context(), w)
}

func (s *Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	entityType := ""
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusInternalServerError)
		return
	}

	entityPrefix := r.FormValue("entityPrefix")
	entityID := r.FormValue("entityID")
	if entityPrefix == "event" {
		entityType = r.FormValue("type")
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Unable to get file from form", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fmt.Println(handler)

	savePath := "./assets/img/upload/"
	savePath = filepath.Join(savePath, entityPrefix)

	if err := os.MkdirAll(savePath, os.ModePerm); err != nil {
		http.Error(w, "Unable to create upload directory", http.StatusInternalServerError)
		return
	}

	fileName := fmt.Sprintf("%s%s", entityID, filepath.Ext(handler.Filename))
	filePath := filepath.Join(savePath, fileName)

	newFile, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		http.Error(w, "Unable to create file on server", http.StatusInternalServerError)
		return
	}

	_, err = io.Copy(newFile, file)
	if err != nil {
		fmt.Printf("Error copying file content: %v\n", err)
		http.Error(w, "Unable to copy file content", http.StatusInternalServerError)
		return
	}

	parsedID := uuid.MustParse(entityID)
	imagePath := helper.ScanForImage(entityID, entityPrefix)

	switch entityPrefix {
	case "event":
		s.EventStore.UpdateEventImage(parsedID, imagePath)

		if entityType == "old" {
			w.Header().Set("HX-Redirect", fmt.Sprintf("/event/%s", entityID))
		} else if entityType == "new" {
			w.Header().Set("HX-Redirect", fmt.Sprintf("/event/%s/add", entityID))
		}
	case "user":
		s.UserStore.UpdateUserImage(parsedID, imagePath)

		w.Header().Set("HX-Redirect", "/user")
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		s.CustomErrorHandler.InternalServerErrorHandler(w, r)
	}

	_, _ = w.Write(jsonResp)
}
