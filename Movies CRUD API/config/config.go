package config

import (
	"os"
)

func GetDBConnectionString() string {
	return os.Getenv("DATABASE_URL")
}
