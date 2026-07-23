package main

import (
	
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"log"
)

func main(){

	authURLStr := os.Getenv("AUTH_SERVICE_URL")
	taskURLStr := os.Getenv("TASK_SERVICE_URL")

	// Set safe fallbacks for local running
	if authURLStr == "" {
		authURLStr = "http://localhost:8080"
	}
	if taskURLStr == "" {
		taskURLStr = "http://localhost:8081"
	}

	// 2. Parse the URLs safely
	targetLogin, err := url.Parse(authURLStr)
	if err != nil {
		log.Fatalf("Invalid Auth Service URL: %v", err)
	}
	proxyLogin := httputil.NewSingleHostReverseProxy(targetLogin)

	targetTask, err := url.Parse(taskURLStr)
	if err != nil {
		log.Fatalf("Invalid Task Service URL: %v", err)
	}
	proxyTask := httputil.NewSingleHostReverseProxy(targetTask)

	// 3. Define the Reverse Proxy Routing Routes
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		r.Host = targetLogin.Host
		proxyLogin.ServeHTTP(w, r)
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
    r.Host = targetLogin.Host
    proxyLogin.ServeHTTP(w, r)
	})

	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		r.Host = targetTask.Host
		proxyTask.ServeHTTP(w, r)
	})

	fmt.Println("Gateway Service is safely starting on port 8000...")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		panic(err)
	}
}