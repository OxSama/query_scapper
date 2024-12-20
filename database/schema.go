// database/schema.go
package database

import (
	"database/sql"
	"fmt"
	"os"
)

// MySQL table creation queries
const (
	mysqlUsersQuery = `
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

	mysqlAuthQuery = `
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

	mysqlLogsQuery = `
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
)

// PostgreSQL table creation queries
const (
	postgresUsersQuery = `
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

	postgresAuthQuery = `
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

	postgresLogsQuery = `
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
)

// CreateTables creates all necessary database tables based on the database type
func CreateTables(db *sql.DB) error {
	dbType := os.Getenv("DB_TYPE")
	var usersQuery, authQuery, logsQuery string

	switch dbType {
	case "mysql":
		usersQuery = mysqlUsersQuery
		authQuery = mysqlAuthQuery
		logsQuery = mysqlLogsQuery
	case "postgres":
		usersQuery = postgresUsersQuery
		authQuery = postgresAuthQuery
		logsQuery = postgresLogsQuery
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	// Start transaction for tables creation
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
