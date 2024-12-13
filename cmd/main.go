package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Environment string
var db *gorm.DB

func init() {
	env := os.Getenv("ENV")
	if env == "" {
		Environment = "development"
		os.Setenv("ENV", Environment)
	} else {
		Environment = env
		os.Setenv("ENV", Environment)
	}

	log.Printf("Running in %s environment", Environment)

	var err error
	dsn := os.Getenv("DATABASE_URL")
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate the User model
	db.AutoMigrate(&User{})

	// Seed initial data if the users table is empty
	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		users := []User{
			{Name: "Davida123"},
			{Name: "Brianabc"},
			{Name: "Jeff"},
		}
		db.Create(&users)
	}
}

// User struct now works with GORM annotations for ORM mapping
type User struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
}

// main function to set up the router and start the server
func main() {
	r := chi.NewRouter()

	// Middleware for logging and recovering from panics
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Define route for getting users
	r.Get("/users", getUsers)

	log.Println("Server is running on port 8080")
	log.Println("Server is running in", os.Getenv("ENV"))

	// Start the HTTP server on port 8080
	http.ListenAndServe(":8080", r)
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
