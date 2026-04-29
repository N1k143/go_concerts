package handlers

import (
	"concerts/internal/repository"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type ReservationHandler struct {
	concertRepo     *repository.ConcertRepo
	seatingRepo     *repository.SeatingRepo
	reservationRepo *repository.ReservationRepo
}

func NewReservationHandler(
	concertRepo *repository.ConcertRepo,
	seatingRepo *repository.SeatingRepo,
	reservationRepo *repository.ReservationRepo,
) *ReservationHandler {
	return &ReservationHandler{
		concertRepo:     concertRepo,
		seatingRepo:     seatingRepo,
		reservationRepo: reservationRepo,
	}
}

type reservationRequest struct {
	ReservationToken *string `json:"reservation_token"`
	Reservations     []struct {
		Row  *int64 `json:"row"`
		Seat *int   `json:"seat"`
	} `json:"reservations"`
	Duration *int `json:"duration"`
}

func (h *ReservationHandler) Reserve(w http.ResponseWriter, r *http.Request) {
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

	var req reservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate duration
	duration := 300
	if req.Duration != nil {
		duration = *req.Duration
	}
	if duration < 1 || duration > 300 {
		writeValidationError(w, map[string]string{
			"duration": "The duration must be between 1 and 300.",
		})
		return
	}

	// Validate reservations array items
	if req.Reservations == nil {
		req.Reservations = []struct {
			Row  *int64 `json:"row"`
			Seat *int   `json:"seat"`
		}{}
	}

	fieldErrors := map[string]string{}
	for _, res := range req.Reservations {
		if res.Row == nil {
			fieldErrors["reservations"] = "The row field is required."
			break
		}
		if res.Seat == nil {
			fieldErrors["reservations"] = "The seat field is required."
			break
		}
	}
	if len(fieldErrors) > 0 {
		writeValidationError(w, fieldErrors)
		return
	}

	// Handle existing reservation token
	var existingReservation *repository.ReservationDB
	var reservationToken string

	if req.ReservationToken != nil && *req.ReservationToken != "" {
		existingReservation, err = h.reservationRepo.GetByToken(*req.ReservationToken)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		if existingReservation == nil {
			writeError(w, http.StatusForbidden, "Invalid reservation token")
			return
		}
		reservationToken = existingReservation.Token
		// Clear seats previously reserved for this reservation+show
		if err := h.seatingRepo.ClearReservationSeats(existingReservation.ID); err != nil {
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	}

	// Get rows for show
	rowMap, err := h.seatingRepo.GetRowsForShow(showID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Validate each reservation seat
	for _, res := range req.Reservations {
		row, ok := rowMap[*res.Row]
		if !ok || row.ShowID != showID {
			writeValidationError(w, map[string]string{
				"reservations": "Seat " + strconv.Itoa(*res.Seat) + " in row " + strconv.FormatInt(*res.Row, 10) + " is invalid.",
			})
			return
		}
		// Check seat exists (number in range)
		seat, err := h.seatingRepo.GetSeat(*res.Row, *res.Seat)
		if err != nil {
			writeValidationError(w, map[string]string{
				"reservations": "Seat " + strconv.Itoa(*res.Seat) + " in row " + strconv.FormatInt(*res.Row, 10) + " is invalid.",
			})
			return
		}
		_ = seat

		var excludeID *int64
		if existingReservation != nil {
			excludeID = &existingReservation.ID
		}
		taken, err := h.seatingRepo.IsSeatTaken(*res.Row, *res.Seat, excludeID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		if taken {
			writeValidationError(w, map[string]string{
				"reservations": "Seat " + strconv.Itoa(*res.Seat) + " in row " + strconv.FormatInt(*res.Row, 10) + " is already taken.",
			})
			return
		}
	}

	expiresAt := time.Now().Add(time.Duration(duration) * time.Second)

	// Create or update reservation
	var reservationID int64
	if existingReservation != nil {
		reservationID = existingReservation.ID
		if err := h.reservationRepo.UpdateExpiry(reservationID, expiresAt); err != nil {
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	} else {
		reservationToken, err = generateToken()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		newRes, err := h.reservationRepo.Create(reservationToken, expiresAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		reservationID = newRes.ID
	}

	// Assign seats
	for _, res := range req.Reservations {
		if err := h.seatingRepo.SetReservation(*res.Row, *res.Seat, &reservationID); err != nil {
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"reserved":          true,
		"reservation_token": reservationToken,
		"reserved_until":    expiresAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
