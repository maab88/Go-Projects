package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func ConnectDB(connString string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}
	return conn, nil
}
