package main

import (
	"gateway/internal/api"
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

	handler := ws.NewHandler(hub, nc)

	upgrader := gws.NewUpgrader(handler, nil)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Printf("Erreur écriture: %v", err)
		}
	})

	r.Post("/api/messages", api.NewMessagesHandler(nc))

	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		socket, err := upgrader.Upgrade(w, r)
		if err != nil {
			log.Printf("Erreur upgrade: %v", err)
			return
		}
		go socket.ReadLoop()
	})

	addr := ":8080"
	log.Printf("Serveur démarré sur http://localhost%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
