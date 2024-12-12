package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
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

	fmt.Printf("DB_TYPE %s", os.Getenv("DB_TYPE"))

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

func createTables(db *sql.DB) error {
	dbType := os.Getenv("DB_TYPE")
	var usersQuery, authQuery, logsQuery string

	switch dbType {
	case "mysql":
		usersQuery = `
			CREATE TABLE IF NOT EXISTS users (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				username VARCHAR(255) NOT NULL UNIQUE,
				email VARCHAR(255) NOT NULL UNIQUE,
				password_hash CHAR(60) NOT NULL,
				full_name VARCHAR(255),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				is_active BOOLEAN DEFAULT TRUE,
				last_login TIMESTAMP,
				INDEX idx_username (username),
				INDEX idx_email (email)
			) ENGINE=InnoDB;
		`
		authQuery = `
			CREATE TABLE IF NOT EXISTS auth_tokens (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				user_id BIGINT NOT NULL,
				token VARCHAR(255) NOT NULL UNIQUE,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				expires_at TIMESTAMP NOT NULL,
				is_revoked BOOLEAN DEFAULT FALSE,
				device_info TEXT,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
				INDEX idx_token (token),
				INDEX idx_user_tokens (user_id, is_revoked)
			) ENGINE=InnoDB;
		`
		logsQuery = `
			CREATE TABLE IF NOT EXISTS auth_logs (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				user_id BIGINT,
				action VARCHAR(50) NOT NULL,
				ip_address VARCHAR(45) NOT NULL,
				user_agent TEXT,
				timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				status_code INT NOT NULL,
				description TEXT,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
				INDEX idx_user_logs (user_id, timestamp),
				INDEX idx_action (action),
				INDEX idx_timestamp (timestamp)
			) ENGINE=InnoDB;
		`
	case "postgres":
		usersQuery = `
			CREATE TABLE IF NOT EXISTS users (
				id BIGSERIAL PRIMARY KEY,
				username VARCHAR(255) NOT NULL UNIQUE,
				email VARCHAR(255) NOT NULL UNIQUE,
				password_hash CHAR(60) NOT NULL,
				full_name VARCHAR(255),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				is_active BOOLEAN DEFAULT TRUE,
				last_login TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_username ON users(username);
			CREATE INDEX IF NOT EXISTS idx_email ON users(email);
		`
		authQuery = `
			CREATE TABLE IF NOT EXISTS auth_tokens (
				id BIGSERIAL PRIMARY KEY,
				user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				token VARCHAR(255) NOT NULL UNIQUE,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				expires_at TIMESTAMP NOT NULL,
				is_revoked BOOLEAN DEFAULT FALSE,
				device_info TEXT
			);
			CREATE INDEX IF NOT EXISTS idx_token ON auth_tokens(token);
			CREATE INDEX IF NOT EXISTS idx_user_tokens ON auth_tokens(user_id, is_revoked);
		`
		logsQuery = `
			CREATE TABLE IF NOT EXISTS auth_logs (
				id BIGSERIAL PRIMARY KEY,
				user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
				action VARCHAR(50) NOT NULL,
				ip_address VARCHAR(45) NOT NULL,
				user_agent TEXT,
				timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				status_code INT NOT NULL,
				description TEXT
			);
			CREATE INDEX IF NOT EXISTS idx_user_logs ON auth_logs(user_id, timestamp);
			CREATE INDEX IF NOT EXISTS idx_action ON auth_logs(action);
			CREATE INDEX IF NOT EXISTS idx_timestamp ON auth_logs(timestamp);
		`
	}

	// Trnxs for tables creation
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Create users table first since it's referenced by others
	_, err = tx.Exec(usersQuery)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create users table: %v", err)
	}

	// Create auth_tokens table
	_, err = tx.Exec(authQuery)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create auth_tokens table: %v", err)
	}

	// Create auth_logs table
	_, err = tx.Exec(logsQuery)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create auth_logs table: %v", err)
	}

	return tx.Commit()
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

	err = createTables(db)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully created all tables!")
}