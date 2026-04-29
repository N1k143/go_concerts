package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type BookingRepo struct {
	db *sqlx.DB
}

func NewBookingRepo(db *sqlx.DB) *BookingRepo {
	return &BookingRepo{db: db}
}

type BookingDB struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Address   string    `db:"address"`
	City      string    `db:"city"`
	Zip       string    `db:"zip"`
	Country   string    `db:"country"`
	CreatedAt time.Time `db:"created_at"`
}

func (r *BookingRepo) Create(name, address, city, zip, country string) (*BookingDB, error) {
	var b BookingDB
	err := r.db.Get(&b, `
		INSERT INTO bookings (name, address, city, zip, country)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, address, city, zip, country, created_at
	`, name, address, city, zip, country)
	return &b, err
}

type TicketDB struct {
	ID        int64     `db:"id"`
	Code      string    `db:"code"`
	BookingID int64     `db:"booking_id"`
	CreatedAt time.Time `db:"created_at"`
	// Joined fields
	BookingName string `db:"booking_name"`
	RowID       int64  `db:"row_id"`
	RowName     string `db:"row_name"`
	SeatNumber  int    `db:"seat_number"`
	ShowID      int64  `db:"show_id"`
	ShowStart   string `db:"show_start"`
	ShowEnd     string `db:"show_end"`
	ConcertID   int64  `db:"concert_id"`
	Artist      string `db:"artist"`
	LocationID  int64  `db:"location_id"`
	LocationName string `db:"location_name"`
}

func (r *BookingRepo) CreateTicket(code string, bookingID int64) (*TicketDB, error) {
	var t TicketDB
	err := r.db.Get(&t, `
		INSERT INTO tickets (code, booking_id) VALUES ($1, $2)
		RETURNING id, code, booking_id, created_at
	`, code, bookingID)
	if err != nil {
		return nil, err
	}
	t.BookingID = bookingID
	return &t, nil
}

func (r *BookingRepo) GetTicketFull(ticketID int64) (*TicketDB, error) {
	var t TicketDB
	err := r.db.Get(&t, `
		SELECT
			tk.id, tk.code, tk.booking_id, tk.created_at,
			bk.name AS booking_name,
			lsr.id AS row_id, lsr.name AS row_name,
			ls.number AS seat_number,
			s.id AS show_id,
			to_char(s.start AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS show_start,
			to_char(s."end" AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS show_end,
			c.id AS concert_id, c.artist,
			l.id AS location_id, l.name AS location_name
		FROM tickets tk
		JOIN bookings bk ON bk.id = tk.booking_id
		JOIN location_seats ls ON ls.ticket_id = tk.id
		JOIN location_seat_rows lsr ON lsr.id = ls.location_seat_row_id
		JOIN shows s ON s.id = lsr.show_id
		JOIN concerts c ON c.id = s.concert_id
		JOIN locations l ON l.id = c.location_id
		WHERE tk.id = $1
	`, ticketID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &t, err
}

func (r *BookingRepo) GetTicketByCodeAndName(code, name string) (*TicketDB, error) {
	var t TicketDB
	err := r.db.Get(&t, `
		SELECT
			tk.id, tk.code, tk.booking_id, tk.created_at,
			bk.name AS booking_name,
			lsr.id AS row_id, lsr.name AS row_name,
			ls.number AS seat_number,
			s.id AS show_id,
			to_char(s.start AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS show_start,
			to_char(s."end" AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS show_end,
			c.id AS concert_id, c.artist,
			l.id AS location_id, l.name AS location_name
		FROM tickets tk
		JOIN bookings bk ON bk.id = tk.booking_id
		JOIN location_seats ls ON ls.ticket_id = tk.id
		JOIN location_seat_rows lsr ON lsr.id = ls.location_seat_row_id
		JOIN shows s ON s.id = lsr.show_id
		JOIN concerts c ON c.id = s.concert_id
		JOIN locations l ON l.id = c.location_id
		WHERE tk.code = $1 AND bk.name = $2
	`, code, name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &t, err
}

func (r *BookingRepo) GetTicketsByBookingID(bookingID int64) ([]TicketDB, error) {
	var tickets []TicketDB
	err := r.db.Select(&tickets, `
		SELECT
			tk.id, tk.code, tk.booking_id, tk.created_at,
			bk.name AS booking_name,
			lsr.id AS row_id, lsr.name AS row_name,
			ls.number AS seat_number,
			s.id AS show_id,
			to_char(s.start AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS show_start,
			to_char(s."end" AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS show_end,
			c.id AS concert_id, c.artist,
			l.id AS location_id, l.name AS location_name
		FROM tickets tk
		JOIN bookings bk ON bk.id = tk.booking_id
		JOIN location_seats ls ON ls.ticket_id = tk.id
		JOIN location_seat_rows lsr ON lsr.id = ls.location_seat_row_id
		JOIN shows s ON s.id = lsr.show_id
		JOIN concerts c ON c.id = s.concert_id
		JOIN locations l ON l.id = c.location_id
		WHERE tk.booking_id = $1
		ORDER BY tk.id
	`, bookingID)
	return tickets, err
}

func (r *BookingRepo) GetTicketByID(ticketID int64) (*TicketDB, error) {
	var t TicketDB
	err := r.db.Get(&t, `
		SELECT
			tk.id, tk.code, tk.booking_id, tk.created_at,
			bk.name AS booking_name,
			lsr.id AS row_id, lsr.name AS row_name,
			ls.number AS seat_number,
			s.id AS show_id,
			to_char(s.start AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS show_start,
			to_char(s."end" AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS show_end,
			c.id AS concert_id, c.artist,
			l.id AS location_id, l.name AS location_name
		FROM tickets tk
		JOIN bookings bk ON bk.id = tk.booking_id
		JOIN location_seats ls ON ls.ticket_id = tk.id
		JOIN location_seat_rows lsr ON lsr.id = ls.location_seat_row_id
		JOIN shows s ON s.id = lsr.show_id
		JOIN concerts c ON c.id = s.concert_id
		JOIN locations l ON l.id = c.location_id
		WHERE tk.id = $1
	`, ticketID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &t, err
}

func (r *BookingRepo) DeleteTicket(ticketID int64) error {
	// Clear the seat assignment first
	_, err := r.db.Exec(`UPDATE location_seats SET ticket_id = NULL WHERE ticket_id = $1`, ticketID)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`DELETE FROM tickets WHERE id = $1`, ticketID)
	return err
}
