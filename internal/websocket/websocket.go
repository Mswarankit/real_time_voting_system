package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Vote struct {
	Option string `json:"option"`
}

type VoteCount struct {
	Option string `json:"option"`
	Count  int    `json:"count"`
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	votes      map[string]int
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		votes:      make(map[string]int),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (c *Client) readPump(h *Hub) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var vote Vote
		if err := json.Unmarshal(message, &vote); err != nil {
			fmt.Println("Error unmarshalling vote:", err)
			continue
		}
		h.votes[vote.Option]++
		h.broadcastVoteCount()
	}
}

// writePump() continuously listens messages
func (c *Client) writePump() {
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
		c.conn.WriteMessage(websocket.CloseMessage, []byte{})
	}
}

// sends each clients current vote counts
func (h *Hub) broadcastVoteCount() {
	for client := range h.clients {
		for option, count := range h.votes {
			voteCount := VoteCount{Option: option, Count: count}
			message, err := json.Marshal(voteCount)
			if err != nil {
				fmt.Println("Error marshalling vote count:", err)
				continue
			}
			client.send <- message
		}
	}
}

func ServeWs(h *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := &Client{conn: conn, send: make(chan []byte, 256)}
	h.register <- client

	go client.writePump()
	go client.readPump(h)
}
