package ws

import (
	"log"
	"strings"
	"sync"

	"github.com/lxzan/gws"
	"github.com/nats-io/nats.go"
)

type Hub struct {
	mu sync.RWMutex

	Rooms map[string]map[string]*gws.Conn
}

func NewHub() *Hub {
	return &Hub{
		Rooms: make(map[string]map[string]*gws.Conn),
	}
}

// StartNatsSubscription écoute les messages NATS pour les redistribuer aux clients WS.
func (h *Hub) StartNatsSubscription(nc *nats.Conn) error {
	// On écoute sur message.broadcast.> pour recevoir les messages destinés à n'importe quelle room.
	// Le joker '>' permet de matcher plusieurs niveaux (ex: message.broadcast.group:123 ou message.broadcast.user.abc)
	_, err := nc.Subscribe("message.broadcast.>", func(m *nats.Msg) {
		parts := strings.Split(m.Subject, ".")
		if len(parts) < 3 {
			return
		}
		// On rejoint toutes les parties après "message.broadcast" au cas où la room contient des points
		roomID := strings.Join(parts[2:], ".")

		log.Printf("[Hub] Message reçu de NATS pour la room %s", roomID)
		h.BroadcastToRoom(roomID, m.Data)
	})

	if err == nil {
		log.Println("[Hub] Abonnement NATS aux sujets message.broadcast.> actif")
	}
	return err
}

func (h *Hub) Join(roomName string, socket *gws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	socketID := socket.RemoteAddr().String()

	if _, exists := h.Rooms[roomName]; !exists {
		h.Rooms[roomName] = make(map[string]*gws.Conn)
		log.Printf("[Hub] Création de la room : %s", roomName)
	}

	h.Rooms[roomName][socketID] = socket
	log.Printf("[Hub] Client %s a rejoint la room %s", socketID, roomName)
}

func (h *Hub) Leave(roomName string, socket *gws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	socketID := socket.RemoteAddr().String()

	if clients, exists := h.Rooms[roomName]; exists {
		delete(clients, socketID)
		log.Printf("[Hub] Client %s a quitté la room %s", socketID, roomName)

		if len(clients) == 0 {
			delete(h.Rooms, roomName)
			log.Printf("[Hub] Room %s supprimée car vide", roomName)
		}
	}
}

func (h *Hub) BroadcastToRoom(roomName string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, exists := h.Rooms[roomName]
	if !exists {
		return
	}

	for _, socket := range clients {
		err := socket.WriteMessage(gws.OpcodeText, payload)
		if err != nil {
			log.Printf("Erreur envoi message : %v", err)
		}
	}
}
