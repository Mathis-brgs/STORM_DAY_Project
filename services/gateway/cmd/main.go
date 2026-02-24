package main

import (
	"gateway/internal/modules/auth"
	"gateway/internal/modules/user"
	"gateway/internal/ws"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lxzan/gws"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Impossible de se connecter à NATS: %v", err)
	}
	defer nc.Close()
	log.Printf("Connecté à NATS sur %s", natsURL)

	hub := ws.NewHub()

	handler := ws.NewHandler(hub, nc)
	upgrader := gws.NewUpgrader(handler, nil)

	authHandler := auth.NewHandler(nc)
	userHandler := user.NewHandler(nc)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Auth Routes
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)
	r.Post("/auth/refresh", authHandler.Refresh)
	r.Post("/auth/logout", authHandler.Logout)

	// User Routes
	r.Get("/users/{id}", userHandler.Get)
	r.Put("/users/{id}", userHandler.Update)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}
	})

	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		socket, err := upgrader.Upgrade(w, r)
		if err != nil {
			log.Printf("Erreur upgrade: %v", err)
			return
		}
		go socket.ReadLoop()
	})

	// Metrics Prometheus
	r.Handle("/metrics", promhttp.Handler())

	addr := ":8080"
	log.Printf("Serveur démarré sur http://localhost%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
