package mdb

import (
	"database/sql"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
)

// EmailEntry represents a single record in the "emails" table.
type EmailEntry struct {
	Id          int64      // Unique identifier (primary key)
	Email       string     // Email address (unique constraint)
	ConfirmedAt *time.Time // Timestamp of confirmation, nullable
	OptOut      bool       // Indicates if the user opted out
}

// TryCreate attempts to create the "emails" table if it doesn't already exist.
func TryCreate(db *sql.DB) {
	_, err := db.Exec(`
	CREATE TABLE emails (
		id             INTEGER PRIMARY KEY,     -- Auto-incrementing primary key
		email          TEXT UNIQUE,             -- Unique email address
		confirmed_at   INTEGER,                 -- Stored as Unix timestamp
		opt_out        INTEGER                  -- Boolean stored as 0 (false) or 1 (true)
	);
	`)
	if err != nil {
		// Cast the error to sqlite3.Error to check for specific SQLite error codes
		if sqlError, ok := err.(sqlite3.Error); ok {
			// Error code 1 = "table already exists"
			if sqlError.Code != 1 {
				log.Fatal(sqlError) // Fatal error if it's something else
			}
		} else {
			// If it's not a recognized SQLite error, still log it
			log.Fatal(err)
		}
	}
}

// emailEntryFromRow converts a database row into an EmailEntry struct.
// This is useful when iterating over query results.
func emailEntryFromRow(row *sql.Rows) (*EmailEntry, error) {
	var (
		id          int64  // To hold the ID value from the row
		email       string // Email string from the row
		confirmedAt int64  // Will be converted to time.Time
		optOut      bool   // Stored as INTEGER in SQLite but mapped to Go bool
	)

	// Extract values from the current row
	err := row.Scan(&id, &email, &confirmedAt, &optOut)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Convert Unix timestamp to time.Time
	t := time.Unix(confirmedAt, 0)

	// Return a pointer to the constructed EmailEntry
	return &EmailEntry{
		Id:          id,
		Email:       email,
		ConfirmedAt: &t,
		OptOut:      optOut,
	}, nil
}

// CreateEmail inserts a new email entry with default values: not confirmed and not opted out.
func CreateEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`
		INSERT INTO emails(email, confirmed_at, opt_out)
		VALUES(?, 0, false)
	`, email)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// GetMail retrieves an email entry from the database by email address.
func GetMail(db *sql.DB, email string) (*EmailEntry, error) {
	rows, err := db.Query(`
		SELECT id, email, confirmed_at, opt_out
		FROM emails
		WHERE email = ?
	`, email)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	// If there's at least one row, convert it to EmailEntry
	for rows.Next() {
		return emailEntryFromRow(rows)
	}
	return nil, nil // No entry found
}

// UpdateEmail upserts an email entry: if it exists, update it; otherwise, insert it.
func UpdateEmail(db *sql.DB, entry EmailEntry) error {
	t := entry.ConfirmedAt.Unix()

	_, err := db.Exec(`
		INSERT INTO emails(email, confirmed_at, opt_out)
		VALUES(?, ?, ?)
		ON CONFLICT(email) DO UPDATE SET
			confirmed_at = excluded.confirmed_at,
			opt_out = excluded.opt_out
	`, entry.Email, t, entry.OptOut)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// DeleteEmail performs a soft delete by setting opt_out = true for the given email.
func DeleteEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`
		UPDATE emails
		SET opt_out = true
		WHERE email = ?
	`, email)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
