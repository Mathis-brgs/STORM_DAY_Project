package main

import (
	"gateway/internal/api"
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
	if err := hub.StartNatsSubscription(nc); err != nil {
		log.Fatalf("Impossible de démarrer l'abonnement NATS : %v", err)
	}

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

	// Message (proxy vers message-service)
	r.Post("/api/messages", api.NewMessagesHandler(nc))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}
	})

	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract Token (Query Param or Header)
		token := r.URL.Query().Get("token")
		if token == "" {
			token = r.Header.Get("Authorization")
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}
		}

		if token == "" {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		// 2. Validate Token via NATS
		valResult, err := auth.ValidateToken(nc, token)
		if err != nil {
			log.Printf("Validation NATS Error: %v", err)
			http.Error(w, "Authentication service unavailable", http.StatusServiceUnavailable)
			return
		}

		if !valResult.Valid {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		// 3. Upgrade et stockage en session
		socket, err := upgrader.Upgrade(w, r)
		if err != nil {
			log.Printf("Erreur upgrade: %v", err)
			return
		}

		// 4. Store user info in socket session
		socket.Session().Store("userId", valResult.User.ID)
		socket.Session().Store("username", valResult.User.Username)

		go socket.ReadLoop()
	})

	addr := ":8080"
	log.Printf("Serveur démarré sur http://localhost%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
