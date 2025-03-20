package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// Nameday represents a mapping of date to names
type Nameday map[string][]string

// InitializeDatabase creates the database and the namedays table
func InitializeDatabase(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS namedays (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		name TEXT NOT NULL
	);
	`
	_, err := db.Exec(query)
	return err
}

// InsertNamedays inserts parsed nameday data into SQLite
func InsertNamedays(db *sql.DB, namedayData Nameday) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO namedays (date, name) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for date, names := range namedayData {
		for _, name := range names {
			_, err = stmt.Exec(date, name)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

// QueryNamedays prints all records in the namedays table
func QueryNamedays(db *sql.DB) {
	rows, err := db.Query("SELECT date, name FROM namedays ORDER BY date")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var date, name string
		if err := rows.Scan(&date, &name); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Date: %s, Name: %s\n", date, name)
	}
}

func main() {
	// Open JSON file
	file, err := os.Open("namedays.json")
	if err != nil {
		log.Fatal("Error opening JSON file:", err)
	}
	defer file.Close()

	// Read JSON file
	jsonData, err := io.ReadAll(file)
	if err != nil {
		log.Fatal("Error reading JSON file:", err)
	}

	// Parse JSON
	var namedayData Nameday
	err = json.Unmarshal(jsonData, &namedayData)
	if err != nil {
		log.Fatal("Error parsing JSON:", err)
	}

	// Initialize SQLite database
	db, err := sql.Open("sqlite3", "./namedays.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = InitializeDatabase(db)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Insert data
	err = InsertNamedays(db, namedayData)
	if err != nil {
		log.Fatal("Failed to insert data:", err)
	}

	// Query and display stored data
	fmt.Println("Namedays in database:")
	QueryNamedays(db)
}
