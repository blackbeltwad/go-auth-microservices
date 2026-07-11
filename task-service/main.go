package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"os"
)

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

func main(){

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
		return
	})

	fmt.Println("Task Service is starting on port 8081...")

	if err := http.ListenAndServe(":8081", nil); err != nil {
		panic(err)
	}


}
