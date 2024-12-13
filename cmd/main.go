package main

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Define your User struct (must match your SQLite and PostgreSQL schema)
type User struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
}

func main() {
	// 1. Connect to the SQLite database
	sqliteDB, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to SQLite database:", err)
	}

	// 2. Connect to the PostgreSQL database
	// Replace the DSN below with your actual PostgreSQL connection string
	postgresDSN := os.Getenv("postgresDSN")
	postgresDB, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL database:", err)
	}

	// Ensure PostgreSQL database has the same schema by migrating
	err = postgresDB.AutoMigrate(&User{})
	if err != nil {
		log.Fatal("Failed to migrate PostgreSQL schema:", err)
	}

	// 3. Retrieve all records from the SQLite database
	var users []User
	if err := sqliteDB.Find(&users).Error; err != nil {
		log.Fatal("Failed to retrieve data from SQLite:", err)
	}

	log.Printf("Found %d users in SQLite", len(users))

	// 4. Insert the data into PostgreSQL, update on conflict
	if len(users) > 0 {
		// Use transaction for safe inserts
		tx := postgresDB.Begin()

		for _, user := range users {
			// Insert each user, updating any records with the same ID
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},              // Define the conflict target (primary key)
				DoUpdates: clause.AssignmentColumns([]string{"name"}), // Update only the 'name' column on conflict
			}).Create(&user).Error; err != nil {
				tx.Rollback() // Rollback in case of any error
				log.Fatal("Failed to insert or update data in PostgreSQL:", err)
			}
		}

		tx.Commit() // Commit the transaction if all goes well
		log.Println("Successfully migrated data to PostgreSQL with upserts!")
	} else {
		log.Println("No users to migrate.")
	}

}
