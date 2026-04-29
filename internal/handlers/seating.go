package handlers

import (
	"concerts/internal/repository"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type SeatingHandler struct {
	concertRepo *repository.ConcertRepo
	seatingRepo *repository.SeatingRepo
}

func NewSeatingHandler(concertRepo *repository.ConcertRepo, seatingRepo *repository.SeatingRepo) *SeatingHandler {
	return &SeatingHandler{concertRepo: concertRepo, seatingRepo: seatingRepo}
}

func (h *SeatingHandler) GetSeating(w http.ResponseWriter, r *http.Request) {
	concertID, err := strconv.ParseInt(chi.URLParam(r, "concert-id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "A concert or show with this ID does not exist")
		return
	}
	showID, err := strconv.ParseInt(chi.URLParam(r, "show-id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "A concert or show with this ID does not exist")
		return
	}

	valid, err := h.concertRepo.ValidateConcertShow(concertID, showID)
	if err != nil || !valid {
		writeError(w, http.StatusNotFound, "A concert or show with this ID does not exist")
		return
	}

	rows, err := h.seatingRepo.GetSeating(showID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rows": rows})
}
