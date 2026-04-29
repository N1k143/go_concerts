package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type ReservationRepo struct {
	db *sqlx.DB
}

func NewReservationRepo(db *sqlx.DB) *ReservationRepo {
	return &ReservationRepo{db: db}
}

type ReservationDB struct {
	ID        int64     `db:"id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
}

func (r *ReservationRepo) Create(token string, expiresAt time.Time) (*ReservationDB, error) {
	var res ReservationDB
	err := r.db.Get(&res, `
		INSERT INTO reservations (token, expires_at) VALUES ($1, $2)
		RETURNING id, token, expires_at
	`, token, expiresAt)
	return &res, err
}

func (r *ReservationRepo) GetByToken(token string) (*ReservationDB, error) {
	var res ReservationDB
	err := r.db.Get(&res, `SELECT id, token, expires_at FROM reservations WHERE token = $1`, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &res, err
}

func (r *ReservationRepo) UpdateExpiry(id int64, expiresAt time.Time) error {
	_, err := r.db.Exec(`UPDATE reservations SET expires_at = $1 WHERE id = $2`, expiresAt, id)
	return err
}

func (r *ReservationRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM reservations WHERE id = $1`, id)
	return err
}
