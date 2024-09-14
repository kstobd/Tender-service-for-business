package db

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4"
	_ "github.com/lib/pq"
)

var conn *pgx.Conn

// Connect устанавливает соединение с базой данных PostgreSQL
func Connect() {
	var err error
	postgresConn := os.Getenv("POSTGRES_CONN")
	log.Printf("Connecting to database at %s", postgresConn)
	conn, err = pgx.Connect(context.Background(), postgresConn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	log.Println("Successfully connected to the database")
}

// GetConnection возвращает текущее соединение с базой данных
func GetConnection() *pgx.Conn {
	return conn
}

// Close закрывает соединение с базой данных
func Close() {
	conn.Close(context.Background())
}

// RunMigrations запускает миграции базы данных
func RunMigrations() {
	log.Println("Starting database migrations...")

	postgresConn := os.Getenv("POSTGRES_CONN") + "?sslmode=disable"
	Db, err := sql.Open("postgres", postgresConn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer Db.Close()

	driver, err := postgres.WithInstance(Db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil {
		log.Fatalf("Could not create migration instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations ran successfully")
}
