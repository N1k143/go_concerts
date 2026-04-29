package repository

import (
	"concerts/internal/models"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type ConcertRepo struct {
	db *sqlx.DB
}

func NewConcertRepo(db *sqlx.DB) *ConcertRepo {
	return &ConcertRepo{db: db}
}

type concertRow struct {
	ID           int64  `db:"id"`
	Artist       string `db:"artist"`
	LocationID   int64  `db:"location_id"`
	LocationName string `db:"location_name"`
}

type showRow struct {
	ID        int64  `db:"id"`
	ConcertID int64  `db:"concert_id"`
	Start     string `db:"start"`
	End       string `db:"end"`
}

func (r *ConcertRepo) ListAll() ([]models.ConcertResponse, error) {
	var concerts []concertRow
	err := r.db.Select(&concerts, `
		SELECT c.id, c.artist, c.location_id, l.name AS location_name
		FROM concerts c
		JOIN locations l ON l.id = c.location_id
		ORDER BY c.id
	`)
	if err != nil {
		return nil, err
	}

	var shows []showRow
	err = r.db.Select(&shows, `
		SELECT id, concert_id,
		       to_char(start AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS start,
		       to_char("end" AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS "end"
		FROM shows ORDER BY concert_id, id
	`)
	if err != nil {
		return nil, err
	}

	showMap := make(map[int64][]models.ShowResponse)
	for _, s := range shows {
		showMap[s.ConcertID] = append(showMap[s.ConcertID], models.ShowResponse{
			ID: s.ID, Start: s.Start, End: s.End,
		})
	}

	result := make([]models.ConcertResponse, 0, len(concerts))
	for _, c := range concerts {
		sh := showMap[c.ID]
		if sh == nil {
			sh = []models.ShowResponse{}
		}
		result = append(result, models.ConcertResponse{
			ID:     c.ID,
			Artist: c.Artist,
			Location: models.LocationResponse{ID: c.LocationID, Name: c.LocationName},
			Shows:  sh,
		})
	}
	return result, nil
}

func (r *ConcertRepo) GetByID(id int64) (*models.ConcertResponse, error) {
	var c concertRow
	err := r.db.Get(&c, `
		SELECT c.id, c.artist, c.location_id, l.name AS location_name
		FROM concerts c
		JOIN locations l ON l.id = c.location_id
		WHERE c.id = $1
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var shows []showRow
	_ = r.db.Select(&shows, `
		SELECT id, concert_id,
		       to_char(start AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS start,
		       to_char("end" AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS "end"
		FROM shows WHERE concert_id = $1 ORDER BY id
	`, id)

	sh := make([]models.ShowResponse, 0, len(shows))
	for _, s := range shows {
		sh = append(sh, models.ShowResponse{ID: s.ID, Start: s.Start, End: s.End})
	}

	return &models.ConcertResponse{
		ID:     c.ID,
		Artist: c.Artist,
		Location: models.LocationResponse{ID: c.LocationID, Name: c.LocationName},
		Shows:  sh,
	}, nil
}

// ValidateConcertShow checks that show belongs to concert; returns (showID found, error)
func (r *ConcertRepo) ValidateConcertShow(concertID, showID int64) (bool, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*) FROM shows WHERE id = $1 AND concert_id = $2
	`, showID, concertID)
	return count > 0, err
}
