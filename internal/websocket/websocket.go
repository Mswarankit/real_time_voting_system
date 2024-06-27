package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"

	"real_time_voting_system/internal/storage"

	"github.com/golang-jwt/jwt/v5"
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
	conn     *websocket.Conn
	send     chan []byte
	username string
}

type Hub struct {
	clients     map[*Client]bool
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	redisClient *storage.RedisClient
	votes       map[string]int
}

func NewHub(redisClient *storage.RedisClient) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		redisClient: redisClient,
		votes:       make(map[string]int),
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

		// Check if the user has already voted
		hasVoted, err := h.redisClient.HasVoted(c.username)
		if err != nil {
			fmt.Println("Error checking vote status:", err)
			continue
		}
		if hasVoted {
			alreadyVotedMessage := struct {
				Message string `json:"message"`
			}{
				Message: "You already voted",
			}
			message, err := json.Marshal(alreadyVotedMessage)
			if err != nil {
				fmt.Println("Error marshalling already voted message:", err)
				continue
			}
			c.send <- message
			continue
		}

		// Set the user as having voted
		if err := h.redisClient.SetVote(c.username); err != nil {
			fmt.Println("Error setting vote status:", err)
			continue
		}

		// Increment the vote count
		h.votes[vote.Option]++
		h.broadcastVoteCounts()
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current WebSocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (h *Hub) broadcastVoteCounts() {
	totalVotes := make(map[string]int)
	for option, count := range h.votes {
		totalVotes[option] = count
	}

	// Determine the winner
	var winner string
	var maxVotes int
	for option, count := range totalVotes {
		if count > maxVotes {
			winner = option
			maxVotes = count
		}
	}

	// Create the message with total votes and the winner
	messageData := struct {
		TotalVotes map[string]int `json:"total_votes"`
		Winner     string         `json:"winner"`
	}{
		TotalVotes: totalVotes,
		Winner:     winner,
	}

	// Marshal the total vote counts and winner into JSON
	message, err := json.Marshal(messageData)
	if err != nil {
		fmt.Println("Error marshalling total vote count:", err)
		return
	}

	// Broadcast the total vote counts and winner to all connected clients
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func ServeWs(h *Hub, w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "token is required", http.StatusUnauthorized)
		return
	}

	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("my_secret_key"), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	username := claims.Subject
	storedToken, err := h.redisClient.GetToken(username)
	if err != nil || storedToken != tokenString {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := &Client{conn: conn, send: make(chan []byte, 256), username: username}
	h.register <- client

	go client.writePump()
	go client.readPump(h)
}
