package helper

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Yatsok/digital-queue/internal/models"
	"github.com/araddon/dateparse"
	"github.com/google/uuid"
)

var (
	nilUserID uuid.UUID
)

func ScanForImage(entityID, entityPrefix string) string {
	imagePath := "/img/placeholder.png"

	fileExt := []string{"png", "jpg", "jpeg"}
	for _, ext := range fileExt {
		fileName := fmt.Sprintf("%s.%s", entityID, ext)
		uploadPath := fmt.Sprintf("assets/img/upload/%s/", entityPrefix)
		filePath := filepath.Join(uploadPath, fileName)

		var err error
		_, err = os.Stat(filePath)
		if err == nil {
			imagePath = fmt.Sprintf("/img/upload/%s/%s.%s", entityPrefix, entityID, ext)
			return imagePath
		}
	}

	return imagePath
}

func CountAvailableTimeSlots(timeSlots []models.TimeSlot) int {
	count := 0

	for _, timeSlot := range timeSlots {
		if timeSlot.UserID == nil {
			count++
		}
	}

	return count
}

func StringToTime(timeStr string) time.Time {
	parsedTime, err := dateparse.ParseAny(timeStr)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return time.Time{}
	}

	return parsedTime
}

func TimeInLocation(timeStr, tz string) time.Time {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		fmt.Println("Error parsing location:", err)
		return time.Time{}
	}

	t, err := time.ParseInLocation("2006-01-02T15:04", timeStr, loc)
	if err != nil {
		t, err = time.Parse("2006-01-02 15:04:05 -0700 MST", timeStr)
		if err != nil {
			fmt.Println("Error parsing time:", err)
			return time.Time{}
		}
	}

	return t.In(loc)
}

func ReturnNilUserID() uuid.UUID {
	return nilUserID
}

func FormatDuration(d time.Duration) string {
	absDuration := d
	if d < 0 {
		absDuration = -d
	}

	hours := int(absDuration.Hours())
	minutes := int(absDuration.Minutes()) % 60

	var result string
	if hours > 0 {
		result = fmt.Sprintf("%d hours", hours)
	}
	if minutes > 0 {
		if hours > 0 {
			result += " and "
		}
		result += fmt.Sprintf("%d minutes", minutes)
	}
	if result == "" {
		return "just now"
	}
	return result
}
