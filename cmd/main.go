package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"l0-wb/internal/cache"
	"l0-wb/internal/db"
	"l0-wb/internal/kafka"
)

const dbMetadata = "postgres://user:password@postgres:5432/orders?sslmode=disable"

func withCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		handler(w, r)
	}
}

func handleOrder(w http.ResponseWriter, r *http.Request, dbConn *sql.DB) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 || pathParts[2] == "" {
		http.Error(w, "Wrong URL format. Use /order/{uid}", http.StatusBadRequest)
		return
	}
	uid := pathParts[2]

	if order, ok := cache.OrdersCache.Get(uid); ok {
		log.Println("Found in cache:", uid)
		json.NewEncoder(w).Encode(order)
		return
	}

	order, err := db.GetFullInfo(dbConn, uid)
	if err != nil {
		log.Printf("Failed to get info from DB: %v\n", err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	cache.OrdersCache.Add(uid, order)
	json.NewEncoder(w).Encode(order)
}

func handleUploadOrder(w http.ResponseWriter, r *http.Request) {
	var order db.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		log.Println("Failed to decode JSON:", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	acceptCh := make(chan struct{})
	go kafka.StartProducer(order, acceptCh)

	<-acceptCh
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Order received")
}

func fillCache(dbConn *sql.DB) {
	cache.InitCache(100)

	rows, err := dbConn.Query("SELECT order_uid FROM orders LIMIT 100")
	if err != nil {
		log.Printf("Failed to preload cache: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			continue
		}
		order, err := db.GetFullInfo(dbConn, uid)
		if err == nil {
			cache.OrdersCache.Add(uid, order)
		}
	}
}

func main() {
	dbConn, err := sql.Open("postgres", dbMetadata)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer dbConn.Close()

	db.RunMigrations(dbConn)
	fillCache(dbConn)

	http.HandleFunc("/order/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		handleOrder(w, r, dbConn)
	}))

	http.HandleFunc("/order/upload", withCORS(handleUploadOrder))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	go kafka.StartConsumer(dbConn)

	log.Printf("Server is listening on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
