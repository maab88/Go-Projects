package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Movie struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	ReleaseDate CustomDate `json:"release_date"`
	Director    Director   `json:"director"`
}

type Director struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

type CustomDate time.Time

func (cd *CustomDate) Scan(src any) error {
	switch v := src.(type) {
	case time.Time:
		*cd = CustomDate(v)
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into CustomDate", v)
	}
}

func (cd CustomDate) Time() time.Time {
	return time.Time(cd)
}

func (cd CustomDate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + cd.Time().Format("2006-01-02") + `"`), nil
}

func (cd *CustomDate) UnmarshalJSON(data []byte) error {
	parsed, err := time.Parse(`"2006-01-02"`, string(data))
	if err != nil {
		return err
	}
	*cd = CustomDate(parsed)
	return nil
}

func connectDB(connString string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func handleDBError(ctx *gin.Context, err error) {
	log.Printf("Database error: %v", err)
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}

func getMoviesHandler(db *pgx.Conn) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rows, err := db.Query(context.Background(), `
			SELECT m.id, m.title, m.release_date, d.id AS director_id, d.first_name, d.last_name
			FROM movies m
			JOIN directors d ON m.director_id = d.id
		`)
		if err != nil {
			handleDBError(ctx, err)
			return
		}
		defer rows.Close()

		var movies []Movie
		for rows.Next() {
			var movie Movie
			var director Director
			if err := rows.Scan(&movie.ID, &movie.Title, &movie.ReleaseDate, &director.ID, &director.FirstName, &director.LastName); err != nil {
				handleDBError(ctx, err)
				return
			}
			movie.Director = director
			movies = append(movies, movie)
		}
		ctx.JSON(http.StatusOK, movies)
	}
}

func getMovieByIDHandler(db *pgx.Conn) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		var movie Movie
		var director Director

		err := db.QueryRow(context.Background(), `
			SELECT m.id, m.title, m.release_date, d.id AS director_id, d.first_name, d.last_name
			FROM movies m
			JOIN directors d ON m.director_id = d.id
			WHERE m.id = $1
		`, id).Scan(&movie.ID, &movie.Title, &movie.ReleaseDate, &director.ID, &director.FirstName, &director.LastName)
		if err != nil {
			handleDBError(ctx, err)
			return
		}

		movie.Director = director
		ctx.JSON(http.StatusOK, movie)
	}
}

func createMovieHandler(db *pgx.Conn) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var movie Movie
		if err := ctx.ShouldBindJSON(&movie); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		err := db.QueryRow(context.Background(), `
			INSERT INTO movies (title, release_date, director_id)
			VALUES ($1, $2, $3)
			RETURNING id
		`, movie.Title, time.Time(movie.ReleaseDate), movie.Director.ID).Scan(&movie.ID)
		if err != nil {
			handleDBError(ctx, err)
			return
		}

		ctx.JSON(http.StatusCreated, movie)
	}
}

func updateMovieHandler(db *pgx.Conn) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		var movie Movie
		if err := ctx.ShouldBindJSON(&movie); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		_, err := db.Exec(context.Background(), `
			UPDATE movies
			SET title = $1, release_date = $2, director_id = $3
			WHERE id = $4
		`, movie.Title, time.Time(movie.ReleaseDate), movie.Director.ID, id)
		if err != nil {
			handleDBError(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Movie updated successfully"})
	}
}

func deleteMovieHandler(db *pgx.Conn) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		_, err := db.Exec(context.Background(), "DELETE FROM movies WHERE id = $1", id)
		if err != nil {
			handleDBError(ctx, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Movie deleted successfully"})
	}
}

func main() {
	connString := "postgres://postgres:test@localhost:5432/movies"
	log.Println("Starting Server on port 8000")

	db, err := connectDB(connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close(context.Background())

	r := gin.Default()

	r.GET("/movies", getMoviesHandler(db))
	r.GET("/movies/:id", getMovieByIDHandler(db))
	r.POST("/movies", createMovieHandler(db))
	r.PUT("/movies/:id", updateMovieHandler(db))
	r.DELETE("/movies/:id", deleteMovieHandler(db))

	r.Run(":8000")
}
