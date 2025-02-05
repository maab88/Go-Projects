package repositories

import (
	"MoviesCRUDAPI/internal/models"
	"context"

	"github.com/jackc/pgx/v5"
)

type DirectorRepository struct {
	DB *pgx.Conn
}

func NewDirectorRepository(db *pgx.Conn) *DirectorRepository {
	return &DirectorRepository{DB: db}
}

func (r *DirectorRepository) GetAllDirectors(ctx context.Context) ([]models.Director, error) {
	rows, err := r.DB.Query(ctx, "SELECT id, first_name, last_name FROM directors")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var directors []models.Director
	for rows.Next() {
		var director models.Director
		if err := rows.Scan(&director.ID, &director.FirstName, &director.LastName); err != nil {
			return nil, err
		}
		directors = append(directors, director)
	}
	return directors, nil
}

func (r *DirectorRepository) GetDirectorByID(ctx context.Context, id int) (models.Director, error) {
	var director models.Director
	err := r.DB.QueryRow(ctx, "SELECT id, first_name, last_name FROM directors WHERE id=$1", id).
		Scan(&director.ID, &director.FirstName, &director.LastName)
	if err != nil {
		return director, err
	}
	return director, nil
}

func (r *DirectorRepository) CreateDirector(ctx context.Context, director models.Director) (int, error) {
	var id int
	err := r.DB.QueryRow(ctx, "INSERT INTO directors (first_name, last_name) VALUES ($1, $2) RETURNING id",
		director.FirstName, director.LastName).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *DirectorRepository) UpdateDirector(ctx context.Context, director models.Director) error {
	_, err := r.DB.Exec(ctx, "UPDATE directors SET first_name=$1, last_name=$2 WHERE id=$3",
		director.FirstName, director.LastName, director.ID)
	return err
}

func (r *DirectorRepository) DeleteDirector(ctx context.Context, id int) error {
	_, err := r.DB.Exec(ctx, "DELETE FROM directors WHERE id=$1", id)
	return err
}
