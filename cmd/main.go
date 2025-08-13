package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"l0-wb/internal/cache"
	"l0-wb/internal/config"
	"l0-wb/internal/kafka"
	"l0-wb/internal/repo"
	"l0-wb/internal/service"

	_ "github.com/lib/pq"
)

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

func main() {
	cfg := config.Load()

	dsn := cfg.PGDSN()
	dbConn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer dbConn.Close()

	time.Sleep(time.Second*3)

	if err := dbConn.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	if err := repo.RunMigrations(dbConn); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	pg := repo.NewPostgresRepo(dbConn)
	if err := cache.InitCache(cfg.CacheSize); err != nil {
		log.Fatalf("cache init: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cache.Preload(ctx, dbConn, cfg.CacheSize)

	prod, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Printf("kafka producer init failed: %v", err)
		prod = nil
	}

	svc := service.NewOrderService(pg, cache.Self(), prod)

	mux := http.NewServeMux()
	mux.HandleFunc("/order/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/order/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "Wrong URL format. Use /order/{uid}", http.StatusBadRequest)
			return
		}
		uid := parts[0]
		o, err := svc.GetByUID(uid)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(o)
	}))
	mux.HandleFunc("/order/upload", withCORS(func(w http.ResponseWriter, r *http.Request) {
		var o service.OrderDTO
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if err := svc.Create(o); err != nil {
			http.Error(w, "save failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintln(w, "accepted")
	}))

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: mux,
	}

	consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, dbConn)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("HTTP listening %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer.Run(ctx)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}

	cancel()

	wg.Wait()
	if prod != nil {
		_ = prod.Close()
	}
	log.Println("bye")
}
