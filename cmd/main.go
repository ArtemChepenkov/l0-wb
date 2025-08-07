package main

import (
	"fmt"
	"net/http"
	"strings"
	"os"
	"log"
	"database/sql"
	data "l0-wb/internal/db"
	"encoding/json"
	"l0-wb/internal/cache"
	"l0-wb/internal/kafka"
)

const (
	metadata = "postgres://user:password@postgres:5432/orders?sslmode=disable"
)

func handleOrder(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	
	w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
        
	if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }
	
	pathParts := strings.Split(r.URL.Path, "/")
	
	if len(pathParts) < 3 {
		http.Error(w, "Wrong URL format", http.StatusBadRequest)
		return
	}

	uid := pathParts[2]

	order, ok := cache.OrdersCache.Get(uid)
	if ok {
		fmt.Println("found in cache")
		
		json.NewEncoder(w).Encode(order)
		return
	}

	order, err := data.GetFullInfo(database, uid)
	if err != nil {
		log.Printf("Failed to get info: %v\n", err)
	}
	cache.OrdersCache.Add(uid, order)

	
	json.NewEncoder(w).Encode(order)
	
}

func fillCache(database *sql.DB) {
	cache.InitCache(100)
	rows, err := database.Query("SELECT order_uid FROM orders LIMIT 100")
	if err != nil {
		log.Printf("Failed to preload cache: %v", err)
	} else {
		for rows.Next() {
			var uid string
			if err := rows.Scan(&uid); err != nil {
				continue
			}
			order, err := data.GetFullInfo(database, uid)
			if err == nil {
				cache.OrdersCache.Add(uid, order)
			}
		}
		rows.Close()
	}
}

func handleUploadOrder(w http.ResponseWriter, r *http.Request) {
	
	w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
	
	var order data.Order
	acceptCh := make(chan struct{})
    err := json.NewDecoder(r.Body).Decode(&order)
    if err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }


    go kafka.StartProducer(order, acceptCh)

	<-acceptCh

    w.WriteHeader(http.StatusAccepted)
    fmt.Fprintln(w, "Order received")
}

func main() {
	database, err := sql.Open("postgres", metadata)
	if err != nil {
		log.Fatalf("Failed to open db: %v\n", err)
	}
	data.RunMigrations(database)

	fillCache(database)

	http.HandleFunc("/order/", func (w http.ResponseWriter, r *http.Request) {
		handleOrder(w, r, database)
	})

	http.HandleFunc("/order/upload", handleUploadOrder)
	

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	go kafka.StartConsumer(database)

	fmt.Printf("\nServer is listening on %s port\n", port)

	http.ListenAndServe(":"+port, nil)
}