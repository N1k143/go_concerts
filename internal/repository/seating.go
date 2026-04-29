package repository

import (
	"concerts/internal/models"
	"time"

	"github.com/jmoiron/sqlx"
)

type SeatingRepo struct {
	db *sqlx.DB
}

func NewSeatingRepo(db *sqlx.DB) *SeatingRepo {
	return &SeatingRepo{db: db}
}

type seatRowDB struct {
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Order  int    `db:"order"`
	ShowID int64  `db:"show_id"`
}

type seatDB struct {
	RowID         int64  `db:"location_seat_row_id"`
	Number        int    `db:"number"`
	ReservationID *int64 `db:"reservation_id"`
	TicketID      *int64 `db:"ticket_id"`
	ExpiresAt     *time.Time `db:"expires_at"`
}

func (r *SeatingRepo) GetSeating(showID int64) ([]models.RowSeatingResponse, error) {
	var rows []seatRowDB
	err := r.db.Select(&rows, `
		SELECT id, name, "order", show_id
		FROM location_seat_rows
		WHERE show_id = $1
		ORDER BY "order"
	`, showID)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return []models.RowSeatingResponse{}, nil
	}

	// Get all seats for this show with reservation expiry
	var seats []seatDB
	err = r.db.Select(&seats, `
		SELECT ls.location_seat_row_id, ls.number, ls.reservation_id, ls.ticket_id,
		       res.expires_at
		FROM location_seats ls
		JOIN location_seat_rows lsr ON lsr.id = ls.location_seat_row_id
		LEFT JOIN reservations res ON res.id = ls.reservation_id
		WHERE lsr.show_id = $1
		ORDER BY ls.location_seat_row_id, ls.number
	`, showID)
	if err != nil {
		return nil, err
	}

	// Build map row_id -> seats
	type seatInfo struct {
		number        int
		reservationID *int64
		ticketID      *int64
		expiresAt     *time.Time
	}
	seatMap := make(map[int64][]seatInfo)
	for _, s := range seats {
		seatMap[s.RowID] = append(seatMap[s.RowID], seatInfo{
			number:        s.Number,
			reservationID: s.ReservationID,
			ticketID:      s.TicketID,
			expiresAt:     s.ExpiresAt,
		})
	}

	now := time.Now()
	result := make([]models.RowSeatingResponse, 0, len(rows))
	for _, row := range rows {
		seatList := seatMap[row.ID]
		total := len(seatList)
		unavailable := []int{}
		for _, s := range seatList {
			if s.ticketID != nil {
				unavailable = append(unavailable, s.number)
			} else if s.reservationID != nil && s.expiresAt != nil && s.expiresAt.After(now) {
				unavailable = append(unavailable, s.number)
			}
		}
		result = append(result, models.RowSeatingResponse{
			ID:   row.ID,
			Name: row.Name,
			Seats: models.SeatsAvailability{
				Total:       total,
				Unavailable: unavailable,
			},
		})
	}
	return result, nil
}

// GetRowsForShow returns rows belonging to a show, keyed by row ID
func (r *SeatingRepo) GetRowsForShow(showID int64) (map[int64]*seatRowDB, error) {
	var rows []seatRowDB
	err := r.db.Select(&rows, `
		SELECT id, name, "order", show_id
		FROM location_seat_rows WHERE show_id = $1
	`, showID)
	if err != nil {
		return nil, err
	}
	m := make(map[int64]*seatRowDB, len(rows))
	for i := range rows {
		m[rows[i].ID] = &rows[i]
	}
	return m, nil
}

// GetSeat gets a seat in a given row by number
func (r *SeatingRepo) GetSeat(rowID int64, number int) (*seatDB, error) {
	var s seatDB
	err := r.db.Get(&s, `
		SELECT ls.location_seat_row_id, ls.number, ls.reservation_id, ls.ticket_id,
		       res.expires_at
		FROM location_seats ls
		LEFT JOIN reservations res ON res.id = ls.reservation_id
		WHERE ls.location_seat_row_id = $1 AND ls.number = $2
	`, rowID, number)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// IsSeatTaken checks if a seat is taken (ticket or active reservation not belonging to given reservationID)
func (r *SeatingRepo) IsSeatTaken(rowID int64, number int, excludeReservationID *int64) (bool, error) {
	now := time.Now()
	var count int
	var err error
	if excludeReservationID != nil {
		err = r.db.Get(&count, `
			SELECT COUNT(*) FROM location_seats ls
			LEFT JOIN reservations res ON res.id = ls.reservation_id
			WHERE ls.location_seat_row_id = $1 AND ls.number = $2
			AND (
				ls.ticket_id IS NOT NULL
				OR (ls.reservation_id IS NOT NULL AND ls.reservation_id != $3 AND res.expires_at > $4)
			)
		`, rowID, number, *excludeReservationID, now)
	} else {
		err = r.db.Get(&count, `
			SELECT COUNT(*) FROM location_seats ls
			LEFT JOIN reservations res ON res.id = ls.reservation_id
			WHERE ls.location_seat_row_id = $1 AND ls.number = $2
			AND (
				ls.ticket_id IS NOT NULL
				OR (ls.reservation_id IS NOT NULL AND res.expires_at > $3)
			)
		`, rowID, number, now)
	}
	return count > 0, err
}

// SetReservation sets reservation_id on a seat
func (r *SeatingRepo) SetReservation(rowID int64, number int, reservationID *int64) error {
	_, err := r.db.Exec(`
		UPDATE location_seats SET reservation_id = $1
		WHERE location_seat_row_id = $2 AND number = $3
	`, reservationID, rowID, number)
	return err
}

// ClearReservationSeats clears all seats assigned to a given reservationID
func (r *SeatingRepo) ClearReservationSeats(reservationID int64) error {
	_, err := r.db.Exec(`
		UPDATE location_seats SET reservation_id = NULL
		WHERE reservation_id = $1
	`, reservationID)
	return err
}

// GetSeatsForReservation returns all seats for a reservation scoped to a show
func (r *SeatingRepo) GetSeatsForReservation(reservationID int64, showID int64) ([]seatDB, error) {
	var seats []seatDB
	err := r.db.Select(&seats, `
		SELECT ls.location_seat_row_id, ls.number, ls.reservation_id, ls.ticket_id, res.expires_at
		FROM location_seats ls
		JOIN location_seat_rows lsr ON lsr.id = ls.location_seat_row_id
		LEFT JOIN reservations res ON res.id = ls.reservation_id
		WHERE ls.reservation_id = $1 AND lsr.show_id = $2
	`, reservationID, showID)
	return seats, err
}

// UpgradeSeatToTicket sets ticket_id and clears reservation_id on a seat
func (r *SeatingRepo) UpgradeSeatToTicket(rowID int64, number int, ticketID int64) error {
	_, err := r.db.Exec(`
		UPDATE location_seats SET ticket_id = $1, reservation_id = NULL
		WHERE location_seat_row_id = $2 AND number = $3
	`, ticketID, rowID, number)
	return err
}

// GetSeatNumberByRowAndReservation returns the seat number
func (r *SeatingRepo) GetSeatsByReservationAndShow(reservationID, showID int64) ([]struct{ RowID int64; Number int }, error) {
	type row struct {
		RowID  int64 `db:"location_seat_row_id"`
		Number int   `db:"number"`
	}
	var rows []row
	err := r.db.Select(&rows, `
		SELECT ls.location_seat_row_id, ls.number
		FROM location_seats ls
		JOIN location_seat_rows lsr ON lsr.id = ls.location_seat_row_id
		WHERE ls.reservation_id = $1 AND lsr.show_id = $2
	`, reservationID, showID)
	if err != nil {
		return nil, err
	}
	result := make([]struct{ RowID int64; Number int }, len(rows))
	for i, r := range rows {
		result[i] = struct{ RowID int64; Number int }{r.RowID, r.Number}
	}
	return result, nil
}
