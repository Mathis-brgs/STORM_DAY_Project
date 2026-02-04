package main

import (
	"gateway/internal/ws"
	"log"
	"net/http"

	"github.com/lxzan/gws"
)

func main() {
	hub := ws.NewHub()
	
	handler := ws.NewHandler(hub)

	upgrader := gws.NewUpgrader(handler, nil)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		socket, err := upgrader.Upgrade(w, r)
		if err != nil {
			log.Printf("Erreur upgrade: %v", err)
			return
		}
		go socket.ReadLoop()
	})

	addr := ":8080"
	log.Printf("Serveur démarré sur http://localhost%s/ws", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}