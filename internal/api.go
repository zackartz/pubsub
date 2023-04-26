package internal

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

// GetRoom function creates a new room with default client value
func GetRoom() *Room {
	return &Room{
		clients: make(map[*websocket.Conn]bool),
	}
}

// Room struct holds a map of websocket.Conn->client
type Room struct {
	clients map[*websocket.Conn]bool
	m sync.Mutex
}

// addClient method adds a client to the room
func (r *Room) addClient(c *websocket.Conn) {
	r.clients[c] = true
}

// removeClient method removes a client from the room
func (r *Room) removeClient(c *websocket.Conn) {
	delete(r.clients, c)
}

// ServeHTTP method implements the http.Handler interface to serve websocket connections
func (rm *Room) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// always allow connections from any origin
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade connection to websocket & return if error
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade err:", err)
		return
	}

	// add client to room
	rm.addClient(ws)
	defer ws.Close()

	for {
		// read message from client, return and remove client if error
		_, m, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				rm.removeClient(ws)
				log.Printf("error: %v", err)
			}
			return
		}

		// send message to all clients
		for k := range rm.clients {
			socket := k
			// run concurrently to ensure no blocking
			go func() {
				rm.m.Lock()
				defer rm.m.Unlock()
				// write message to client, remove client if error
				err = socket.WriteMessage(websocket.TextMessage, m)
				if err != nil {
					log.Printf("error: %v", err)
					rm.removeClient(socket)
					socket.Close()
				}
			}()
		}
	}
}
