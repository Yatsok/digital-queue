package helper

import (
	"net/http"

	"github.com/google/uuid"
)

func CalculatePagination(currentPage, totalPages int) []int {
	const maxVisiblePages = 5

	startPage := currentPage - maxVisiblePages/2
	if startPage < 1 {
		startPage = 1
	}

	endPage := startPage + maxVisiblePages - 1
	if endPage > totalPages {
		endPage = totalPages
		startPage = endPage - maxVisiblePages + 1
		if startPage < 1 {
			startPage = 1
		}
	}

	var pagination []int
	for i := startPage; i <= endPage; i++ {
		pagination = append(pagination, i)
	}

	return pagination
}

func CheckOwnership(userID, ownerID uuid.UUID) bool {
	return userID == ownerID
}

func OwnershipRedirect(w http.ResponseWriter, r *http.Request, ownershipStatus bool) {
	if !ownershipStatus {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}
