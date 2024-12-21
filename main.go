package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"sql-scapper/database"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type LogEntry struct {
	UserID      int64
	Action      string
	IPAddress   string
	UserAgent   string
	Timestamp   time.Time
	StatusCode  int
	Description string
}

func getDBConnection() (*sql.DB, error) {
	dbType := os.Getenv("DB_TYPE")
	var db *sql.DB
	var err error

	switch dbType {
	case "mysql":
		connStr := fmt.Sprintf(
			"%s:%s@tcp(%s:3306)/%s?parseTime=true",
			os.Getenv("MYSQL_USER"),
			os.Getenv("MYSQL_PASSWORD"),
			os.Getenv("MYSQL_HOST"),
			os.Getenv("MYSQL_DATABASE"),
		)
		db, err = sql.Open("mysql", connStr)
	case "postgres":
		connStr := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("PG_HOST"),
			os.Getenv("PG_USER"),
			os.Getenv("PG_PASSWORD"),
			os.Getenv("PG_DATABASE"),
		)
		db, err = sql.Open("postgres", connStr)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	if err != nil {
		return nil, err
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func insertAuthLog(db *sql.DB, entry LogEntry) error {
	query := `
		INSERT INTO auth_logs 
		(user_id, action, ip_address, user_agent, timestamp, status_code, description)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	if os.Getenv("DB_TYPE") == "postgres" {
		query = strings.Replace(query, "?", "$1", 1)
		for i := 2; i <= 7; i++ {
			query = strings.Replace(query, "?", fmt.Sprintf("$%d", i), 1)
		}
	}

	_, err := db.Exec(query,
		entry.UserID,
		entry.Action,
		entry.IPAddress,
		entry.UserAgent,
		entry.Timestamp,
		entry.StatusCode,
		entry.Description,
	)
	return err
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	fmt.Printf("DB_TYPE: %s\n", os.Getenv("DB_TYPE"))
	fmt.Printf("MYSQL_HOST: %s\n", os.Getenv("MYSQL_HOST"))
	fmt.Printf("MYSQL_USER: %s\n", os.Getenv("MYSQL_USER"))
	fmt.Printf("MYSQL_DATABASE: %s\n", os.Getenv("MYSQL_DATABASE"))

	time.Sleep(10 * time.Second)

	db, err := getDBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Successfully connected to %s database!\n", os.Getenv("DB_TYPE"))

	err = database.CreateTables(db)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully created all tables!")
}
