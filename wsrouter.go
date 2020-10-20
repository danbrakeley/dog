package dog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

type WsMsg struct {
	Router *WsRouter
	Client *WsClient
	Msg    []byte
}

type routerEventType uint8

const (
	retClose routerEventType = iota
	retBroadcast
)

type wsRouterEvent struct {
	Type    routerEventType
	Payload json.RawMessage
}

type WsRouter struct {
	indexHash    string // let's a connected client know when their cached web page is out of date
	msgHandler   func(WsMsg)
	clients      map[*WsClient]bool
	chRegister   chan *WsClient
	chUnregister chan *WsClient
	chEvent      chan wsRouterEvent
	wgPump       sync.WaitGroup
	chDead       chan struct{}
}

// NewWsRouter creates a new WsRouter type
func NewWsRouter(indexHash string) *WsRouter {
	return &WsRouter{
		indexHash:    indexHash,
		clients:      make(map[*WsClient]bool),
		chRegister:   make(chan *WsClient),
		chUnregister: make(chan *WsClient),
		chEvent:      make(chan wsRouterEvent),
		chDead:       make(chan struct{}),
	}
}

// serveWs upgrades http requests to websocket connections
func (r *WsRouter) serveWs(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error upgrading connection to websocket: %v\n", err)
		return
	}

	c := CreateWsClient(conn, r)
	select {
	case <-r.chDead:
		c.BeginShutDown()
	case r.chRegister <- c:
	}
}

func (r *WsRouter) SetClientMsgHandler(fn func(WsMsg)) {
	r.msgHandler = fn
}

func (r *WsRouter) handleClientMessage(c *WsClient, b []byte) {
	if r.msgHandler != nil {
		r.msgHandler(WsMsg{Router: r, Client: c, Msg: b})
	}
}

func (r *WsRouter) handleClientError(c *WsClient, err error) {
	fmt.Fprintf(os.Stderr, "error in websocket router: %v\n", err)
}

func (r *WsRouter) handleClientShutdown(c *WsClient) {
	select {
	case <-r.chDead:
	case r.chUnregister <- c:
	}
}

// Start our websocket router, accepting various requests
func (r *WsRouter) Start() {
	r.wgPump.Add(1)
	go func() {
		isShuttingDown := false
		defer func() {
			close(r.chDead)
			r.wgPump.Done()
		}()
		for {
		outerSelect:
			select {
			case c := <-r.chRegister:
				if isShuttingDown {
					c.BeginShutDown()
					fmt.Printf("===== WsRouter register called during shutdown for client %v (ignored)\n", c)
					break outerSelect
				}
				fmt.Printf("===== WsRouter registering client %v\n", c)
				r.clients[c] = true
			case c := <-r.chUnregister:
				fmt.Printf("===== WsRouter ungeristering client %v\n", c)
				if _, ok := r.clients[c]; ok {
					delete(r.clients, c)
				}
				fmt.Printf("===== WsRouter len(clients) = %d\n", len(r.clients))
				if len(r.clients) == 0 && isShuttingDown {
					fmt.Println("===== WsRouter last client unregistered, so halt main loop")
					c.WaitForShutDown()
					return
				}
			case evt := <-r.chEvent:
				switch evt.Type {
				case retClose:
					if isShuttingDown {
						break outerSelect
					}
					isShuttingDown = true
					for client := range r.clients {
						fmt.Printf("===== WsRouter telling client %v to shut down\n", client)
						client.BeginShutDown()
					}
				case retBroadcast:
					if isShuttingDown {
						break outerSelect
					}
					for client := range r.clients {
						client.Send(evt.Payload)
					}
				default:
					panic(fmt.Errorf("WsRouter encountered unknown ret %d", evt.Type))
				}
			}
		}
	}()
}

func (r *WsRouter) Broadcast(msg json.RawMessage) {
	r.chEvent <- wsRouterEvent{Type: retBroadcast, Payload: msg}
}

func (r *WsRouter) BeginShutdown() {
	r.chEvent <- wsRouterEvent{Type: retClose, Payload: nil}
}

func (r *WsRouter) WaitForShutdown() {
	r.wgPump.Wait()
}
