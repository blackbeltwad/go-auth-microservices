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
	
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		username TEXT NOT NULL,
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

		var username string
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if subValue, exists := claims["sub"].(string); exists {
				username = subValue
			}
		}

		if username == "" {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		
		//---------------------------------------------------------------
		// HANDLES GET
		
		if r.Method == http.MethodGet {
			// Query all tasks where username matches the token's subject
			query := `SELECT id, title, done FROM tasks WHERE username = $1`
			rows, err := db.QueryContext(r.Context(), query, username)
			if err != nil {
				log.Printf("Database query failed: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			tasks := []Task{}
			for rows.Next() {
				var t Task
				if err := rows.Scan(&t.Id, &t.Title, &t.Done); err != nil {
					log.Printf("Row scan failed: %v", err)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				tasks = append(tasks, t)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(tasks)
			return
		}

		//----------------------------------------------------------------
		// HANDLES POST
		if r.Method == http.MethodPost{
			var task Task

			err := json.NewDecoder(r.Body).Decode(&task)
			if err != nil {
				http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
				return
			}

			insertQuery := `INSERT INTO tasks (username, done, title) VALUES ($1, $2, $3)`
			_, err = db.ExecContext(r.Context(), insertQuery, username, task.Done, task.Title)

			if err != nil {
				http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"message": "Task created successfully"})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		
	})

	fmt.Println("Task Service is starting on port 8081...")

	if err := http.ListenAndServe(":8081", nil); err != nil {
		panic(err)
	}


}
