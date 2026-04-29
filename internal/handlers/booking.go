package handlers

import (
	"concerts/internal/models"
	"concerts/internal/repository"
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type BookingHandler struct {
	concertRepo     *repository.ConcertRepo
	seatingRepo     *repository.SeatingRepo
	reservationRepo *repository.ReservationRepo
	bookingRepo     *repository.BookingRepo
}

func NewBookingHandler(
	concertRepo *repository.ConcertRepo,
	seatingRepo *repository.SeatingRepo,
	reservationRepo *repository.ReservationRepo,
	bookingRepo *repository.BookingRepo,
) *BookingHandler {
	return &BookingHandler{
		concertRepo:     concertRepo,
		seatingRepo:     seatingRepo,
		reservationRepo: reservationRepo,
		bookingRepo:     bookingRepo,
	}
}

type bookingRequest struct {
	ReservationToken *string `json:"reservation_token"`
	Name             *string `json:"name"`
	Address          *string `json:"address"`
	City             *string `json:"city"`
	Zip              *string `json:"zip"`
	Country          *string `json:"country"`
}

func (h *BookingHandler) Book(w http.ResponseWriter, r *http.Request) {
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

	var req bookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	fieldErrors := map[string]string{}
	if req.ReservationToken == nil || *req.ReservationToken == "" {
		fieldErrors["reservation_token"] = "The reservation token field is required."
	}
	if req.Name == nil || *req.Name == "" {
		fieldErrors["name"] = "The name field is required."
	}
	if req.Address == nil || *req.Address == "" {
		fieldErrors["address"] = "The address field is required."
	}
	if req.City == nil || *req.City == "" {
		fieldErrors["city"] = "The city field is required."
	}
	if req.Zip == nil || *req.Zip == "" {
		fieldErrors["zip"] = "The zip field is required."
	}
	if req.Country == nil || *req.Country == "" {
		fieldErrors["country"] = "The country field is required."
	}
	if len(fieldErrors) > 0 {
		writeValidationError(w, fieldErrors)
		return
	}

	// Check reservation token
	reservation, err := h.reservationRepo.GetByToken(*req.ReservationToken)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if reservation == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get seats for this reservation+show
	seats, err := h.seatingRepo.GetSeatsByReservationAndShow(reservation.ID, showID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if len(seats) == 0 {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Create booking
	booking, err := h.bookingRepo.Create(*req.Name, *req.Address, *req.City, *req.Zip, *req.Country)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Create tickets and upgrade seats
	var ticketResponses []models.TicketResponse
	for _, seat := range seats {
		code, err := generateTicketCode()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		ticket, err := h.bookingRepo.CreateTicket(code, booking.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		if err := h.seatingRepo.UpgradeSeatToTicket(seat.RowID, seat.Number, ticket.ID); err != nil {
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Get full ticket details
		fullTicket, err := h.bookingRepo.GetTicketFull(ticket.ID)
		if err != nil || fullTicket == nil {
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		ticketResponses = append(ticketResponses, toTicketResponse(fullTicket))
	}

	// Delete reservation
	_ = h.reservationRepo.Delete(reservation.ID)

	if ticketResponses == nil {
		ticketResponses = []models.TicketResponse{}
	}
	writeJSON(w, http.StatusCreated, map[string]any{"tickets": ticketResponses})
}

func generateTicketCode() (string, error) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 10)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		code[i] = chars[n.Int64()]
	}
	return string(code), nil
}

func toTicketResponse(t *repository.TicketDB) models.TicketResponse {
	return models.TicketResponse{
		ID:        t.ID,
		Code:      t.Code,
		Name:      t.BookingName,
		CreatedAt: t.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		Row: models.RowShortResponse{
			ID:   t.RowID,
			Name: t.RowName,
		},
		Seat: t.SeatNumber,
		Show: models.ShowFullResponse{
			ID:    t.ShowID,
			Start: t.ShowStart,
			End:   t.ShowEnd,
			Concert: models.ConcertInTicket{
				ID:     t.ConcertID,
				Artist: t.Artist,
				Location: models.LocationResponse{
					ID:   t.LocationID,
					Name: t.LocationName,
				},
			},
		},
	}
}

// TicketHandler handles GET /api/v1/tickets and POST /api/v1/tickets/{ticket-id}/cancel
type TicketHandler struct {
	bookingRepo *repository.BookingRepo
	seatingRepo *repository.SeatingRepo
}

func NewTicketHandler(bookingRepo *repository.BookingRepo, seatingRepo *repository.SeatingRepo) *TicketHandler {
	return &TicketHandler{bookingRepo: bookingRepo, seatingRepo: seatingRepo}
}

type ticketLookupRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (h *TicketHandler) GetTickets(w http.ResponseWriter, r *http.Request) {
	var req ticketLookupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Find ticket by code + name
	t, err := h.bookingRepo.GetTicketByCodeAndName(req.Code, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if t == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Return all tickets from same booking
	tickets, err := h.bookingRepo.GetTicketsByBookingID(t.BookingID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	resp := make([]models.TicketResponse, 0, len(tickets))
	for i := range tickets {
		resp = append(resp, toTicketResponse(&tickets[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"tickets": resp})
}

func (h *TicketHandler) CancelTicket(w http.ResponseWriter, r *http.Request) {
	ticketID, err := strconv.ParseInt(chi.URLParam(r, "ticket-id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "A ticket with this ID does not exist")
		return
	}

	var req ticketLookupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ticket, err := h.bookingRepo.GetTicketByID(ticketID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if ticket == nil {
		writeError(w, http.StatusNotFound, "A ticket with this ID does not exist")
		return
	}

	if ticket.Code != req.Code || ticket.BookingName != req.Name {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.bookingRepo.DeleteTicket(ticketID); err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
