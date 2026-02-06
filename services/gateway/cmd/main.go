package main

import (
	"gateway/internal/ws"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lxzan/gws"
)

func main() {
	hub := ws.NewHub()

	handler := ws.NewHandler(hub)

	upgrader := gws.NewUpgrader(handler, nil)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

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
