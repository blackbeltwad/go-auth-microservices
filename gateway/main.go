package main

import (
	
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main(){

	targetLogin, _ := url.Parse("http://localhost:8080")
	proxyLogin := httputil.NewSingleHostReverseProxy(targetLogin)

	targetTask, _ := url.Parse("http://localhost:8081")
	proxyTask := httputil.NewSingleHostReverseProxy(targetTask)

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {

		r.Host = targetLogin.Host
		proxyLogin.ServeHTTP(w,r)
		
		
		})
	
	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		
		r.Host = targetTask.Host
		proxyTask.ServeHTTP(w,r)
		
		})


	fmt.Println("Welcome to the gateway on port 8000...")
	
	// This keeps the server running and listening on port 8000
	if err := http.ListenAndServe(":8000", nil); err != nil {
		panic(err)
	}
}