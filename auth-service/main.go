package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"log"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"golang.org/x/crypto/bcrypt"
)
var db *sql.DB

func getJWTKey() []byte {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        return []byte("temporary_local_development_fallback_key")
    }
    return []byte(secret)
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

	// 5. Create our user table automatically if it doesn't exist yet
	query := `
	CREATE TABLE IF NOT EXISTS users (
    	id SERIAL PRIMARY KEY,
    	username VARCHAR(50) UNIQUE NOT NULL,
    	password_hash TEXT NOT NULL,
    	created_at TIMESTAMP WITH TIME DEFAULT CURRENT_TIMESTAMP
	);`
	

	_, err = db.Exec(query)
	if err != nil {
		log.Fatalf("Error creating tasks table: %v", err)
	}

	fmt.Println("Successfully connected to PostgreSQL and verified tables!")
}


type LoginRequest struct{
		Username string `json:"username"`
		Password string `json:"password"`
	}



func main() {

	initDB()
	
	// This creates a simple web route
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
			return
		}
		
		var ask LoginRequest

		err := json.NewDecoder(r.Body).Decode(&ask)

		if err != nil {
			http.Error(w, "Incorrect JSON format", http.StatusBadRequest)
			return
		}

		query := `SELECT password_hash FROM users WHERE username = $1`

		var storedHash string

		err = db.QueryRow(query, ask.Username).Scan(&storedHash)

		if err != nil {
			if err == sql.ErrNoRows {
				log.Println("Username not found")
				http.Error(w, "Bad Password", http.StatusUnauthorized)
				return
			}
			log.Printf("Database query error: %v", err)
    		http.Error(w, "Internal server error", http.StatusInternalServerError)
    		return
		}

		val := CheckPasswordHash(ask.Password, storedHash)

		if val == true{

			tokenString, err := generation(ask.Username)
			if err != nil{
				http.Error(w, "Failed to generate token", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
		    fmt.Fprint(w, tokenString)
			return
		} else{
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			
		}
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {

			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
				return	
			}

			var ask LoginRequest
			// Decode if its a proper json into ask
			err := json.NewDecoder(r.Body).Decode(&ask)

			if err != nil {
				http.Error(w, "Incorrect JSON format", http.StatusBadRequest)
				return
			}

			hash, err := HashPassword(ask.Password)
			if err != nil {
        		http.Error(w, "Failed to process password", http.StatusInternalServerError)
        		return
    		}

			query := `INSERT INTO users (username, password_hash) VALUES ($1,$2)`

			_ , err = db.Exec(query, ask.Username, hash)

			if err != nil {
       		 	http.Error(w, "Username already exists or database error", http.StatusConflict)
        		return
    		}
			w.WriteHeader(http.StatusCreated)
   			json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
			
		})



	fmt.Println("Auth Service is starting on port 8080...")
	
	// This keeps the server running and listening on port 8080
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func generation(user string) (string, error){

	claim := make(jwt.MapClaims)

	claim["sub"] = user
	claim["exp"] = time.Now().Add(15 * time.Minute).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	tokenString, err := token.SignedString(getJWTKey())

	if err != nil{
		return "", err
	}

	return tokenString, nil
	
}

func HashPassword(password string) (string, error) {
	// GenerateFromPassword
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 13)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}


func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
