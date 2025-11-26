package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func getPostgresDSN() (string, error) {
	err := load()
	if err != nil {
		return "", fmt.Errorf("error loading .env file: %w", err)
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
	return dsn, nil
}

func cleanupDatabase(db *sqlx.DB) error {
	_, err := db.Exec(`
        TRUNCATE TABLE pull_request_reviewers CASCADE;
        TRUNCATE TABLE pull_requests CASCADE;
        TRUNCATE TABLE users CASCADE;
        TRUNCATE TABLE teams CASCADE;
    `)
	return err
}

func load() error {
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Join(filepath.Dir(filename), "../..")
	envPath := filepath.Join(rootDir, ".env")

	return godotenv.Load(envPath)
}
