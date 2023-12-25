package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Yatsok/digital-queue/internal/controllers"
	"github.com/Yatsok/digital-queue/internal/database"
	"github.com/Yatsok/digital-queue/internal/handlers"
	"github.com/Yatsok/digital-queue/internal/helper"
	"github.com/Yatsok/digital-queue/internal/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	*http.Server
	port                  int
	db                    database.Service
	Router                *mux.Router
	connectedClientsMutex sync.Mutex
	connectedClients      map[*websocket.Conn]struct{}
	currentUserMap        map[string]*websocket.Conn
	UserStore             controllers.UserStore
	UserHandler           *handlers.UserHandler
	AuthStore             controllers.AuthStore
	AuthHandler           *handlers.AuthHandler
	AuthMiddlewareStore   controllers.AuthMiddlewareStore
	AuthMiddlewareHandler *handlers.AuthMiddlewareHandler
	EventStore            controllers.EventStore
	EventHandler          *handlers.EventHandler
	SlotStore             controllers.SlotStore
	SlotHandler           *handlers.SlotHandler
	NotificationStore     controllers.NotificationStore
	NotificationHandler   *handlers.NotificationHandler
	WebSocketServer       *http.Server
	CustomErrorHandler    *handlers.CustomErrorHandler
}

func NewServer() *Server {
	r := mux.NewRouter()
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	userStore := controllers.NewUserService(database.New().GetDB())
	eventStore := controllers.NewEventService(database.New().GetDB())
	slotStore := controllers.NewSlotService(database.New().GetDB())
	notificationStore := controllers.NewNotificationService(database.New().GetDB())
	authStore := controllers.NewAuthService(userStore)
	authMiddlewareStore := controllers.NewAuthMiddlewareService(authStore)

	s := &Server{
		port:                  port,
		db:                    database.New(),
		Router:                r,
		connectedClientsMutex: sync.Mutex{},
		connectedClients:      make(map[*websocket.Conn]struct{}),
		currentUserMap:        make(map[string]*websocket.Conn),
		CustomErrorHandler:    handlers.NewCustomErrorHandler(),
		UserHandler:           handlers.NewUserHandler(userStore, authStore, eventStore, slotStore),
		EventHandler:          handlers.NewEventHandler(userStore, slotStore, eventStore, authStore),
		SlotHandler:           handlers.NewSlotHandler(userStore, slotStore, eventStore, authStore),
		NotificationHandler:   handlers.NewNotificationHandler(notificationStore, authStore),
		AuthHandler:           handlers.NewAuthHandler(authStore),
		AuthMiddlewareHandler: handlers.NewAuthMiddlewareHandler(authMiddlewareStore, authStore),
		UserStore:             userStore,
		EventStore:            eventStore,
		SlotStore:             slotStore,
		NotificationStore:     notificationStore,
		AuthStore:             authStore,
		AuthMiddlewareStore:   authMiddlewareStore,
	}

	s.Server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      s.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return s
}

func (s *Server) CheckUpcomingTimeSlots(threshold time.Duration) {
	currentTime := time.Now().UTC()

	timeSlots, err := s.SlotStore.GetUpcomingTimeSlots(currentTime, threshold)
	if err != nil {
		fmt.Printf("Error fetching upcoming time slots: %v\n", err)
		return
	}

	for _, timeSlot := range timeSlots {
		timeUntilStart := timeSlot.StartTime.Sub(currentTime)

		if timeUntilStart <= threshold {
			fmt.Printf("Upcoming time slot ID: %s, Start Time: %s\n", timeSlot.ID, timeSlot.StartTime)
			if timeSlot.UserID != nil {
				var user models.User
				err := s.db.GetDB().First(&user, "id = ?", timeSlot.UserID).Error
				if err != nil {
					fmt.Println("User not found for notification service")
				}

				if user.IsSubscribed {
					nt := models.NewNotification(
						*timeSlot.UserID,
						timeSlot.EventName,
						helper.TimeInLocation(timeSlot.StartTime.String(), user.Timezone).Sub(time.Now().UTC()),
						false,
					)

					s.SendMessageToUser(nt)

					s.NotificationStore.CreateNotification(nt)
				}
			}
		}
	}
}

func (s *Server) SendNotificationToClients(notification string) {
	// Iterate over connected clients and send the notification
	for client := range s.connectedClients {
		err := client.WriteJSON(map[string]interface{}{
			"event":        "notification",
			"notification": notification,
		})
		if err != nil {
			fmt.Println("Error sending notification to client:", err)
		}
	}
}

func (s *Server) SendMessageToUser(nt *models.Notification) {
	notification := fmt.Sprintf("%v until your queue time: %s", helper.FormatDuration(nt.TimeLeft), nt.EventName)

	if conn, ok := s.currentUserMap[nt.UserID.String()]; ok {
		err := conn.WriteJSON(map[string]interface{}{
			"event":        "notification",
			"notification": notification,
		})
		if err != nil {
			fmt.Printf("Error sending message to user %s: %v\n", nt.UserID.String(), err)
		}
	} else {
		fmt.Printf("User with ID %s is not connected\n", nt.UserID.String())
	}
}

func (s *Server) DeleteExpiredTimeSlots() {
	currentTime := time.Now().UTC()

	expiredTimeSlots, err := s.SlotStore.GetExpiredTimeSlots(currentTime)
	if err != nil {
		fmt.Printf("Error fetching expired time slots: %v\n", err)
		return
	}

	for _, timeSlot := range expiredTimeSlots {
		fmt.Printf("Deleting expired time slot ID: %s, End Time: %s\n", timeSlot.ID, timeSlot.EndTime)
		err := s.SlotStore.DeleteTimeSlot(timeSlot.ID)
		if err != nil {
			fmt.Printf("Error deleting time slot ID %s: %v\n", timeSlot.ID, err)
		}
	}
}
