package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	log.Println("Server is running on http://0.0.0.0:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
