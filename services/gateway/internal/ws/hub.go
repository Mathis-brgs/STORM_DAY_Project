package ws

import (
	"log"
	"sync"

	"github.com/lxzan/gws"
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
		_ = socket.WriteMessage(gws.OpcodeText, payload)
	}
}
