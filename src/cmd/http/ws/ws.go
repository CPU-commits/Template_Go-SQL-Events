package ws

import (
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	httpUtils "github.com/CPU-commits/Template_Go-EventDriven/src/cmd/http/utils"
	"github.com/CPU-commits/Template_Go-EventDriven/src/settings"
	"github.com/CPU-commits/Template_Go-EventDriven/src/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var settingsData = settings.GetSettings()

// WS Config
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return utils.Includes(strings.Split(settingsData.CORS_DOMAINS, ","), r.Header.Get("Origin"))
	},
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type inbound struct {
	Type      string `json:"type"`
	Namespace string `json:"namespace"`
	Payload   any    `json:"payload"`
}

// Hub
type Message struct {
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	Payload   any    `json:"payload"`
}

type Client struct {
	ID        string
	Namespace string
	Send      chan Message
	Rooms     []string
}

type NamespaceHandler interface {
	ClientHasAccess(c *gin.Context, rooms []string) error
	GetClientID(c *gin.Context) string
	Namespace() string
}

type Hub struct {
	mu                sync.RWMutex
	namespaces        map[string]map[*Client]bool
	namespacesRooms   map[string]map[string][]*Client
	clients           map[string]*Client
	register          chan *Client
	unregister        chan *Client
	broadcast         chan Message
	namespaceHandlers map[string]NamespaceHandler
}

func NewHubWS() *Hub {
	return &Hub{
		namespaces:        make(map[string]map[*Client]bool),
		register:          make(chan *Client),
		unregister:        make(chan *Client),
		broadcast:         make(chan Message),
		namespaceHandlers: make(map[string]NamespaceHandler),
		clients:           make(map[string]*Client),
		namespacesRooms:   make(map[string]map[string][]*Client),
	}
}

func (h *Hub) RegisterHandler(nsHandler NamespaceHandler) {
	h.namespaceHandlers[nsHandler.Namespace()] = nsHandler
}

func (h *Hub) Send(namespace string, clientID string, typeName string, payload any) {
	if c, ok := h.clients[clientID]; ok {
		c.Send <- Message{
			Namespace: namespace,
			Type:      typeName,
			Payload:   payload,
		}
	}
}

func (h *Hub) SendToRoom(namespace string, room string, typeName string, payload any) {
	if clients, ok := h.namespacesRooms[namespace][room]; ok {
		for _, c := range clients {
			c.Send <- Message{
				Namespace: namespace,
				Type:      typeName,
				Payload:   payload,
			}
		}
	}
}

func containsClient(clients []*Client, c *Client) bool {
	return slices.Contains(clients, c)
}

func (h *Hub) addClientToRooms(c *Client) {
	if _, ok := h.namespacesRooms[c.Namespace]; !ok {
		h.namespacesRooms[c.Namespace] = make(map[string][]*Client)
	}

	nsRooms := h.namespacesRooms[c.Namespace]
	for _, room := range c.Rooms {
		clients := nsRooms[room]
		if !containsClient(clients, c) {
			nsRooms[room] = append(clients, c)
		}
	}
}

func (h *Hub) removeClientFromRooms(c *Client) {
	nsRooms, ok := h.namespacesRooms[c.Namespace]
	if !ok {
		return
	}

	for _, room := range c.Rooms {
		if clients, ok := nsRooms[room]; ok {
			for i := range clients {
				if clients[i] == c {
					clients[i] = clients[len(clients)-1]
					clients = clients[:len(clients)-1]
					break
				}
			}
			if len(clients) == 0 {
				delete(nsRooms, room)
			} else {
				nsRooms[room] = clients
			}
		}
	}

	if len(nsRooms) == 0 {
		delete(h.namespacesRooms, c.Namespace)
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if _, ok := h.namespaces[c.Namespace]; !ok {
				h.namespaces[c.Namespace] = make(map[*Client]bool)
			}
			if _, ok := h.namespacesRooms[c.Namespace]; !ok {
				h.namespacesRooms[c.Namespace] = make(map[string][]*Client)
			}
			h.namespaces[c.Namespace][c] = true
			h.addClientToRooms(c)
			h.clients[c.ID] = c
			h.mu.Unlock()
		case c := <-h.unregister:
			h.mu.Lock()
			id := c.ID

			h.removeClientFromRooms(c)
			if namespace, ok := h.namespaces[c.Namespace]; ok {
				if _, present := namespace[c]; present {
					delete(namespace, c)
					close(c.Send)
					if len(namespace) == 0 {
						delete(h.namespaces, c.Namespace)
					}
				}
			}
			delete(h.clients, id)
			h.mu.Unlock()
		case msg := <-h.broadcast:
			h.mu.RLock()
			if namespace, ok := h.namespaces[msg.Namespace]; ok {
				for cli := range namespace {
					select {
					case cli.Send <- msg:
					default:
						h.mu.RUnlock()
						h.mu.Lock()

						h.removeClientFromRooms(cli)

						delete(namespace, cli)
						close(cli.Send)
						if len(namespace) == 0 {
							delete(h.namespaces, msg.Namespace)
						}
						h.mu.Unlock()
						h.mu.RLock()
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// WS
func ServeWS(h *Hub, namespace string) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler, ok := h.namespaceHandlers[namespace]
		if !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, httpUtils.ProblemDetails{
				Title: "No namespace handled",
			})
			return
		}
		rooms := c.QueryArray("rooms")

		if err := handler.ClientHasAccess(c, rooms); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httpUtils.ProblemDetails{
				Title: err.Error(),
			})
			return
		}

		id := handler.GetClientID(c)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		client := &Client{
			ID:        id,
			Namespace: namespace,
			Send:      make(chan Message, 256),
			Rooms:     rooms,
		}
		h.register <- client
		// Reader
		go func() {
			defer func() {
				h.unregister <- client
				conn.Close()
			}()
			conn.SetReadLimit(1 << 20)
			conn.SetReadDeadline(time.Now().Add(pongWait))
			conn.SetPongHandler(func(appData string) error {
				conn.SetReadDeadline(time.Now().Add(pongWait))
				return nil
			})
			for {
				var in inbound
				if err := conn.ReadJSON(&in); err != nil {
					return
				}
				switch in.Type {
				case "join":
					h.unregister <- client
					client.Namespace = in.Namespace
					h.register <- client
				case "message":
					h.broadcast <- Message{
						Type:      "message",
						Namespace: client.Namespace,
						Payload: map[string]any{
							"from": client.ID,
							"data": in.Payload,
						},
					}
				}
			}
		}()
		// Writer
		go func() {
			ticker := time.NewTicker(pingPeriod)
			defer func() {
				ticker.Stop()
				conn.Close()
			}()
			for {
				select {
				case msg, ok := <-client.Send:
					_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
					if !ok {
						conn.WriteMessage(websocket.CloseMessage, []byte{})
						return
					}
					if err := conn.WriteJSON(msg); err != nil {
						return
					}
				case <-ticker.C:
					_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
					if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						return
					}
				}
			}
		}()
	}
}
