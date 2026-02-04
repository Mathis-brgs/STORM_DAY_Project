package ws

import (
	"log"
	"sync"

	"github.com/lxzan/gws"
)

type Hub struct {
	// Le Mutex protège notre map (c'est le gardien de l'armoire)
	mu sync.RWMutex

	// Rooms : Notre structure à deux niveaux
	// Niveau 1 (Clé string) : Le nom de la room (ex: "salon")
	// Niveau 2 (Valeur map) : La liste des gens dans cette room
	Rooms map[string]map[string]*gws.Conn
}

func NewHub() *Hub {
	return &Hub{
		Rooms: make(map[string]map[string]*gws.Conn),
	}
}

// Join : Un client rejoint une room spécifique
func (h *Hub) Join(roomName string, socket *gws.Conn) {
	// 1. On verrouille l'écriture (STOP tout le monde, je modifie le registre !)
	h.mu.Lock()
	defer h.mu.Unlock() // On déverrouille quoi qu'il arrive à la fin de la fonction

	// 2. On récupère l'ID unique du socket (IP:Port)
	socketID := socket.RemoteAddr().String()

	// 3. Si la room n'existe pas encore, on la crée (on ouvre un nouveau tiroir)
	if _, exists := h.Rooms[roomName]; !exists {
		h.Rooms[roomName] = make(map[string]*gws.Conn)
		log.Printf("[Hub] Création de la room : %s", roomName)
	}

	// 4. On ajoute le client dans la room
	h.Rooms[roomName][socketID] = socket
	log.Printf("[Hub] Client %s a rejoint la room %s", socketID, roomName)
}

// Leave : Un client quitte une room spécifique
func (h *Hub) Leave(roomName string, socket *gws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	socketID := socket.RemoteAddr().String()

	// On vérifie que la room existe
	if clients, exists := h.Rooms[roomName]; exists {
		// On supprime le client de cette map
		delete(clients, socketID)
		log.Printf("[Hub] Client %s a quitté la room %s", socketID, roomName)

		// Optionnel : Si la room est vide, on supprime la room pour libérer la mémoire
		if len(clients) == 0 {
			delete(h.Rooms, roomName)
			log.Printf("[Hub] Room %s supprimée car vide", roomName)
		}
	}
}

// BroadcastToRoom : Envoie un message uniquement aux gens d'une room
func (h *Hub) BroadcastToRoom(roomName string, payload []byte) {
	// 1. On verrouille en LECTURE seulement (RLock)
	// Plusieurs broadcasts peuvent se faire en même temps, tant que personne ne modifie la liste des rooms
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 2. On vérifie si la room existe
	clients, exists := h.Rooms[roomName]
	if !exists {
		return // La room n'existe pas, on ne fait rien
	}

	// 3. On envoie à tous les membres de cette room
	for _, socket := range clients {
		// On ignore les erreurs d'écriture pour l'instant (fire and forget)
		_ = socket.WriteMessage(gws.OpcodeText, payload)
	}
}
