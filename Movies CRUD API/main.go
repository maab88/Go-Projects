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
	str := string(data)
	parsed, err := time.Parse(`"2006-01-02"`, str)
	if err != nil {
		return err
	}
	*cd = CustomDate(parsed)
	return nil
}

func connectDB() (*pgx.Conn, error) {
	connString := "postgres://postgres:test@localhost:5432/movies"
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func getMovies(ctx *gin.Context, db *pgx.Conn) {
	rows, err := db.Query(context.Background(), `
		SELECT m.id, m.title, m.release_date, d.id AS director_id, d.first_name, d.last_name
		FROM movies m
		JOIN directors d ON m.director_id = d.id
	`)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var movie Movie
		var director Director
		err := rows.Scan(&movie.ID, &movie.Title, &movie.ReleaseDate, &director.ID, &director.FirstName, &director.LastName)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		movie.Director = director
		movies = append(movies, movie)
	}
	ctx.JSON(http.StatusOK, movies)
}

func getMovieByID(ctx *gin.Context, db *pgx.Conn) {
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
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	movie.Director = director
	ctx.JSON(http.StatusOK, movie)
}

func createMovie(ctx *gin.Context, db *pgx.Conn) {
	var movie Movie
	if err := ctx.ShouldBindJSON(&movie); err != nil {
		log.Printf("JSON binding error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input, check JSON format and types"})
		return
	}

	// Insert into the database
	err := db.QueryRow(context.Background(), `
		INSERT INTO movies (title, release_date, director_id)
		VALUES ($1, $2, $3)
		RETURNING id
	`, movie.Title, time.Time(movie.ReleaseDate), movie.Director.ID).Scan(&movie.ID)

	if err != nil {
		log.Printf("Database error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, movie)
}

func updateMovie(ctx *gin.Context, db *pgx.Conn) {
	id := ctx.Param("id")
	var movie Movie
	if err := ctx.ShouldBindJSON(&movie); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Update the database
	_, err := db.Exec(context.Background(), `
		UPDATE movies
		SET title = $1, release_date = $2, director_id = $3
		WHERE id = $4
	`, movie.Title, time.Time(movie.ReleaseDate), movie.Director.ID, id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Movie updated successfully"})
}

func deleteMovie(ctx *gin.Context, db *pgx.Conn) {
	id := ctx.Param("id")

	// Delete the movie
	_, err := db.Exec(context.Background(), "DELETE FROM movies WHERE id = $1", id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Movie deleted successfully"})
}

func main() {

	fmt.Printf("Starting Server at port 8000\n")

	db, err := connectDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v\n", err)
	}
	defer db.Close(context.Background())

	//Initialize Router
	r := gin.Default()

	// API routes
	r.GET("/movies", func(ctx *gin.Context) {
		getMovies(ctx, db)
	})

	r.GET("/movies/:id", func(ctx *gin.Context) {
		getMovieByID(ctx, db)
	})

	r.POST("/movies", func(ctx *gin.Context) {
		createMovie(ctx, db)
	})

	r.PUT("/movies/:id", func(ctx *gin.Context) {
		updateMovie(ctx, db)
	})

	r.DELETE("/movies/:id", func(ctx *gin.Context) {
		deleteMovie(ctx, db)
	})

	fmt.Printf("Stoping Server at port 8000\n")
	r.Run(":8000")
}
