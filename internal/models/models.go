package models

import "time"

type Location struct {
	ID   int64  `db:"id"   json:"id"`
	Name string `db:"name" json:"name"`
}

type Concert struct {
	ID         int64    `db:"id"          json:"id"`
	Artist     string   `db:"artist"      json:"artist"`
	LocationID int64    `db:"location_id" json:"-"`
	Location   Location `json:"location"`
	Shows      []Show   `json:"shows"`
}

type Show struct {
	ID        int64     `db:"id"         json:"id"`
	ConcertID int64     `db:"concert_id" json:"-"`
	Start     time.Time `db:"start"      json:"start"`
	End       time.Time `db:"end"        json:"end"`
}

type Reservation struct {
	ID        int64     `db:"id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
}

type Booking struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Address   string    `db:"address"`
	City      string    `db:"city"`
	Zip       string    `db:"zip"`
	Country   string    `db:"country"`
	CreatedAt time.Time `db:"created_at"`
}

type Ticket struct {
	ID        int64     `db:"id"`
	Code      string    `db:"code"`
	BookingID int64     `db:"booking_id"`
	CreatedAt time.Time `db:"created_at"`
}

type LocationSeatRow struct {
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Order  int    `db:"order"`
	ShowID int64  `db:"show_id"`
}

type LocationSeat struct {
	ID                int64  `db:"id"`
	LocationSeatRowID int64  `db:"location_seat_row_id"`
	Number            int    `db:"number"`
	ReservationID     *int64 `db:"reservation_id"`
	TicketID          *int64 `db:"ticket_id"`
}

// ===== API response types =====

type ConcertResponse struct {
	ID       int64            `json:"id"`
	Artist   string           `json:"artist"`
	Location LocationResponse `json:"location"`
	Shows    []ShowResponse   `json:"shows"`
}

type LocationResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ShowResponse struct {
	ID    int64  `json:"id"`
	Start string `json:"start"`
	End   string `json:"end"`
}

type RowSeatingResponse struct {
	ID    int64           `json:"id"`
	Name  string          `json:"name"`
	Seats SeatsAvailability `json:"seats"`
}

type SeatsAvailability struct {
	Total       int   `json:"total"`
	Unavailable []int `json:"unavailable"`
}

type TicketResponse struct {
	ID        int64              `json:"id"`
	Code      string             `json:"code"`
	Name      string             `json:"name"`
	CreatedAt string             `json:"created_at"`
	Row       RowShortResponse   `json:"row"`
	Seat      int                `json:"seat"`
	Show      ShowFullResponse   `json:"show"`
}

type RowShortResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ShowFullResponse struct {
	ID      int64            `json:"id"`
	Start   string           `json:"start"`
	End     string           `json:"end"`
	Concert ConcertInTicket  `json:"concert"`
}

type ConcertInTicket struct {
	ID       int64            `json:"id"`
	Artist   string           `json:"artist"`
	Location LocationResponse `json:"location"`
}
