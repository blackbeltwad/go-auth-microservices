package main

import (
	"github.com/joho/godotenv"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"os"
)

func getJWTKey() []byte {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        return []byte("temporary_local_development_fallback_key")
    }
    return []byte(secret)
}
type LoginRequest struct{
		Username string `json:"username"`
		Password string `json:"password"`
	}



func main() {

	godotenv.Load()
	

	// This creates a simple web route for our "Ticket Booth"
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

		if ask.Username == "admin" && ask.Password == "banana"{
			tokenString, err := generation(ask.Username)

			if err != nil{
				http.Error(w, "Failed to generate token", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, tokenString)
			return
		} 

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
