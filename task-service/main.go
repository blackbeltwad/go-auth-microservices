package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"log"

	"github.com/golang-jwt/jwt/v5"
	

	_ "github.com/lib/pq"
)

var db *sql.DB

func getJWTKey() []byte {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        return []byte("temporary_local_development_fallback_key")
    }
    return []byte(secret)
}


type Task struct{
	Id int `json:"id"`
	Done bool `json:"done"`
	Title string `json:"title"`
}

func initDB(){
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	// 3. Open the database connection pool
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// 4. Ping the database to make sure it's actually alive and responding
	if err = db.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}

	// 5. Create our tasks table automatically if it doesn't exist yet
	// Note: Postgres uses SERIAL for auto-incrementing IDs instead of AUTOINCREMENT
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		done BOOLEAN NOT NULL DEFAULT false
	);`

	_, err = db.Exec(query)
	if err != nil {
		log.Fatalf("Error creating tasks table: %v", err)
	}

	fmt.Println("Successfully connected to PostgreSQL and verified tables!")
}

func main(){
	initDB()
	 

	//Task to test validity
	var mock Task
	mock.Id = 258678
	mock.Done = false
	mock.Title = "Coding"

	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		authorized := r.Header.Get("Authorization")

		if authorized == ""{
			http.Error(w, "Token does not exist", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authorized, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization Format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    		return getJWTKey(), nil
		})

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		

		
		jsonData, err := json.Marshal(mock)
		if err != nil {
			fmt.Printf("Error marshaling to JSON: %v\n", err)
		}
		fmt.Fprint(w, string(jsonData))
		
	})

	fmt.Println("Task Service is starting on port 8081...")

	if err := http.ListenAndServe(":8081", nil); err != nil {
		panic(err)
	}


}
