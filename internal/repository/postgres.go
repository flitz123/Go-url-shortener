package repository

import (
	"database/sql"
	"go-url-shortener/internal/models"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) Save(url *models.URL) error {
	_, err := r.db.Exec(
		"INSERT INTO urls (code, original, clicks) VALUES ($1, $2, $3)",
		url.Code, url.Original, url.Clicks)
	return err
}

func (r *PostgresRepo) Get(code string) (*models.URL, error) {
	row := r.db.QueryRow("SELECT code, original, clicks FROM urls WHERE code = $1", code)

	var u models.URL
	err := row.Scan(&u.Code, &u.Original, &u.Clicks)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err // or custom error
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepo) Increment(code string) error {
	_, err := r.db.Exec("UPDATE urls SET clicks = clicks + 1 WHERE code = $1", code)
	return err
}
