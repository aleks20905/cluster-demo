package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User struct now works with GORM annotations for ORM mapping
type User struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
}

var db *gorm.DB

// init function to initialize the database
func init() {
	var err error

	// Get the database file path from the RAILWAY_VOLUME_MOUNT_PATH environment variable
	volumePath := os.Getenv("RAILWAY_VOLUME_MOUNT_PATH")
	if volumePath == "" {
		log.Fatal("RAILWAY_VOLUME_MOUNT_PATH environment variable is not set", volumePath)
	}

	// Define the path to the SQLite database file within the volume
	dbFilePath := fmt.Sprintf("%s/db.sqlite", volumePath)

	// Open SQLite connection using the full file path
	db, err = gorm.Open(sqlite.Open(dbFilePath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate the User struct to create users table
	db.AutoMigrate(&User{})

	// Optionally seed initial users if the table is empty
	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		// Add some initial users
		users := []User{
			{Name: "Davida123"},
			{Name: "Brianabc"},
			{Name: "Jeff"},
		}
		db.Create(&users)
	}
}

func main() {
	uh := userHandler{}
	http.Handle("/users", uh)
	log.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}

type userHandler struct{}

func (uh userHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getUsers(w, r)
	default:
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// getUsers now retrieves data from the database using GORM
func getUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	// Fetch all users from the database
	result := db.Find(&users)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Convert the users to JSON
	b, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the response as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
