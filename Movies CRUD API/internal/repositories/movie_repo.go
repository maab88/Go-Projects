package repositories

import (
	"MoviesCRUDAPI/internal/models"
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type MovieRepository struct {
	DB *pgx.Conn
}

func NewMovieRepository(db *pgx.Conn) *MovieRepository {
	return &MovieRepository{DB: db}
}

func (r *MovieRepository) GetAllMovies(ctx context.Context) ([]models.Movie, error) {
	rows, err := r.DB.Query(ctx, "SELECT id, title, release_date, director_id FROM movies")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var movie models.Movie
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.ReleaseDate, &movie.DirectorID); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	return movies, nil
}

func (r *MovieRepository) GetMovieByID(ctx context.Context, id int) (models.Movie, error) {
	var movie models.Movie
	err := r.DB.QueryRow(ctx, "SELECT id, title, release_date, director_id FROM movies WHERE id=$1", id).
		Scan(&movie.ID, &movie.Title, &movie.ReleaseDate, &movie.DirectorID)
	if err != nil {
		return movie, err
	}
	return movie, nil
}

func (r *MovieRepository) CreateMovie(ctx context.Context, movie models.Movie) (int, error) {
	var id int
	err := r.DB.QueryRow(ctx, "INSERT INTO movies (title, release_date, director_id) VALUES ($1, $2, $3) RETURNING id",
		movie.Title, time.Time(movie.ReleaseDate), movie.DirectorID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *MovieRepository) UpdateMovie(ctx context.Context, movie models.Movie) error {
	_, err := r.DB.Exec(ctx, "UPDATE movies SET title=$1, release_date=$2, director_id=$3 WHERE id=$4",
		movie.Title, time.Time(movie.ReleaseDate), movie.DirectorID, movie.ID)
	return err
}

func (r *MovieRepository) DeleteMovie(ctx context.Context, id int) error {
	_, err := r.DB.Exec(ctx, "DELETE FROM movies WHERE id=$1", id)
	return err
}
