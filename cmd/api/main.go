package main

import (
	"time"

	"github.com/Yatsok/digital-queue/internal/server"
)

func main() {
	server := server.NewServer()

	go server.StartWebSocketServer()

	// Goroutine to check upcoming time slots
	go func() {
		checkInterval := time.Minute
		ticker := time.NewTicker(checkInterval)

		for range ticker.C {
			server.CheckUpcomingTimeSlots(11 * time.Minute)
		}
	}()

	// Goroutine to check and delete expired time slots
	go func() {
		deleteInterval := 10 * time.Minute
		deleteTicker := time.NewTicker(deleteInterval)

		for range deleteTicker.C {
			server.DeleteExpiredTimeSlots()
		}
	}()

	err := server.ListenAndServe()
	if err != nil {
		panic("cannot start server")
	}
}
