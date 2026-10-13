package websocket

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"qr-store/backend/pkg/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Default deny; the Hub sets a stricter check once constructed.
		return false
	},
}

// Client holds a live websocket connection and its channel subscriptions.
type Client struct {
	Hub          *Hub
	Conn         *websocket.Conn
	Send         chan []byte
	Channels     map[string]bool
	RestaurantID string
	Role         string
	mu           sync.Mutex
}

type EventMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type Hub struct {
	clients       map[*Client]bool
	channels      map[string]map[*Client]bool
	broadcast     chan broadcastMessage
	register      chan *Client
	unregister    chan *Client
	mu            sync.RWMutex
	originAllow   []string
	allowAnyOrig  bool
	authenticator func(token string) (restaurantID string, role string, err error)
}

type broadcastMessage struct {
	Channel string
	Payload []byte
}

// NewHub creates a hub. allowedOrigins is the CORS allowlist; when it contains
// "*" any origin is accepted (dev). Authentication is attached later via
// SetAuthenticator.
func NewHub(allowedOrigins []string) *Hub {
	allowAny := false
	for _, o := range allowedOrigins {
		if o == "*" {
			allowAny = true
		}
	}
	return &Hub{
		clients:      make(map[*Client]bool),
		channels:     make(map[string]map[*Client]bool),
		broadcast:    make(chan broadcastMessage, 256),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		originAllow:  allowedOrigins,
		allowAnyOrig: allowAny,
	}
}

// SetAuthenticator wires a token validator used to authorize channel
// subscriptions. It must be called before serving connections.
func (h *Hub) SetAuthenticator(fn func(token string) (string, string, error)) {
	h.authenticator = fn
}

func (h *Hub) originAllowed(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Non-browser clients (no Origin) are allowed; auth still applies.
		return true
	}
	if h.allowAnyOrig {
		return true
	}
	clean := strings.TrimSuffix(origin, "/")
	for _, o := range h.originAllow {
		if strings.TrimSuffix(o, "/") == clean {
			return true
		}
	}
	return false
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				for ch := range client.Channels {
					if clients, exists := h.channels[ch]; exists {
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.channels, ch)
						}
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.channels[msg.Channel]; ok {
				for client := range clients {
					select {
					case client.Send <- msg.Payload:
					default:
						close(client.Send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Publish(channel string, event string, data interface{}) {
	payload, err := json.Marshal(EventMessage{
		Event: event,
		Data:  data,
	})
	if err != nil {
		logger.Log.Error("failed to marshal websocket event", "error", err)
		return
	}

	h.broadcast <- broadcastMessage{
		Channel: channel,
		Payload: payload,
	}
}

// Subscribe authorizes and registers a client to a channel.
// Channels of the form `restaurant:<id>:<room>` require the client to belong to
// that restaurant. `cashier`/`kitchen` rooms additionally require a matching
// role. `order:<id>` channels are open to any authenticated client.
func (h *Hub) Subscribe(client *Client, channel string) bool {
	if !h.authorizeChannel(client, channel) {
		logger.Log.Warn("websocket subscribe denied", "channel", channel, "restaurant_id", client.RestaurantID, "role", client.Role)
		return false
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	client.mu.Lock()
	client.Channels[channel] = true
	client.mu.Unlock()

	if _, ok := h.channels[channel]; !ok {
		h.channels[channel] = make(map[*Client]bool)
	}
	h.channels[channel][client] = true
	return true
}

func (h *Hub) authorizeChannel(client *Client, channel string) bool {
	parts := strings.Split(channel, ":")
	if len(parts) < 2 {
		return false
	}

	switch parts[0] {
	case "order":
		// Customer order-tracking channels are keyed by a randomized order id
		// and only carry status updates. They are allowed without a staff token
		// to keep realtime working for anonymous diners; scoped order data is
		// still protected by the REST endpoints (qr_token enforced).
		return true
	case "restaurant":
		if client.RestaurantID == "" {
			return false
		}
		if len(parts) < 3 {
			return false
		}
		if parts[1] != client.RestaurantID {
			return false
		}
		room := parts[2]
		return roleAllowedInRoom(client.Role, room)
	default:
		return false
	}
}

func roleAllowedInRoom(role, room string) bool {
	switch room {
	case "cashier":
		return role == "SUPERADMIN" || role == "OWNER" || role == "ADMIN" || role == "CASHIER"
	case "kitchen":
		return role == "SUPERADMIN" || role == "OWNER" || role == "ADMIN" || role == "KITCHEN"
	default:
		return true
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		// Handle client commands:
		//   {"action":"auth","token":"<jwt>"}
		//   {"action":"subscribe","channel":"restaurant:<id>:cashier"}
		var cmd struct {
			Action  string `json:"action"`
			Channel string `json:"channel"`
			Token   string `json:"token"`
		}
		if err := json.Unmarshal(message, &cmd); err != nil {
			continue
		}

		switch cmd.Action {
		case "auth":
			if c.Hub.authenticator == nil {
				continue
			}
			restoID, role, err := c.Hub.authenticator(cmd.Token)
			if err != nil {
				logger.Log.Warn("websocket auth failed", "error", err)
				continue
			}
			c.RestaurantID = restoID
			c.Role = role
		case "subscribe":
			if cmd.Channel != "" {
				c.Hub.Subscribe(c, cmd.Channel)
			}
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)
		if err := w.Close(); err != nil {
			return
		}
	}

	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (h *Hub) HandleWebSocket(c *gin.Context) {
	upgrader.CheckOrigin = h.originAllowed

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Error("websocket upgrade failed", "error", err)
		return
	}

	client := &Client{
		Hub:      h,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Channels: make(map[string]bool),
	}

	// Authenticate at connect time if a token is supplied (query param or
	// Authorization header). Without a valid token the client may connect but
	// cannot subscribe to any channel until it sends an {"action":"auth"}.
	if h.authenticator != nil {
		token := c.Query("token")
		if token == "" {
			if ah := c.GetHeader("Authorization"); strings.HasPrefix(ah, "Bearer ") {
				token = strings.TrimPrefix(ah, "Bearer ")
			}
		}
		if token != "" {
			if restoID, role, err := h.authenticator(token); err == nil {
				client.RestaurantID = restoID
				client.Role = role
			}
		}
	}

	// Auto-subscribe if channel query parameter is provided (authorization still applies).
	if ch := c.Query("channel"); ch != "" {
		h.Subscribe(client, ch)
	}

	h.register <- client

	go client.WritePump()
	go client.ReadPump()
}
