package dog

import (
	"fmt"
	"net/http"
	"os"
	"sync"
)

// RET is Router Event Type
type RET uint8

const (
	RETClose RET = iota
	RETBroadcast
)

type WsRouterEvent struct {
	Type    RET
	Payload []byte
}

type WsRouter struct {
	chShuttingDown chan struct{}
	clients        map[*WsClient]bool
	register       chan *WsClient
	unregister     chan *WsClient
	chEvent        chan WsRouterEvent
	wg             sync.WaitGroup
}

// NewWsRouter creates a new WsRouter type
func NewWsRouter() *WsRouter {
	return &WsRouter{
		chShuttingDown: make(chan struct{}),
		clients:        make(map[*WsClient]bool),
		register:       make(chan *WsClient),
		unregister:     make(chan *WsClient),
		chEvent:        make(chan WsRouterEvent),
	}
}

func (r *WsRouter) IsShuttingDown() bool {
	select {
	case <-r.chShuttingDown:
		return true
	default:
		return false
	}
}

// ServeWs upgrades http requests to websocket connections
func (r *WsRouter) ServeWs(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error upgrading connection to websocket: %v", err)
		return
	}

	CreateWsClient(conn, r)
}

func (r *WsRouter) HandleClientMessage(c *WsClient, mt int, b []byte) {
	// TODO: do something with client messages
	fmt.Printf("!!! CLIENT MESSAGE [%d]: %s", mt, string(b))
}

func (r *WsRouter) HandleClientError(c *WsClient, err error) {
	fmt.Fprintf(os.Stderr, "error in websocket router: %v", err)
}

// Start our websocket router, accepting various requests
func (r *WsRouter) Start() {
	r.wg.Add(1)
	go func() {
		isShuttingDown := false
		defer r.wg.Done()
		for {
		outerSelect:
			select {
			case client := <-r.register:
				if isShuttingDown {
					// TODO: tell new client to shut down?
					fmt.Printf("===== WsRouter register called during shutdown for client %v (ignored)\n", client)
					break outerSelect
				}
				fmt.Printf("===== WsRouter registering client %v\n", client)
				r.clients[client] = true
			case client := <-r.unregister:
				fmt.Printf("===== WsRouter ungeristering client %v\n", client)
				if _, ok := r.clients[client]; ok {
					delete(r.clients, client)
				}
				fmt.Printf("===== WsRouter len(clients) = %d\n", len(r.clients))
				if len(r.clients) == 0 && isShuttingDown {
					fmt.Println("===== WsRouter last client unregistered, so halt main loop")
					client.WaitForShutDown()
					return
				}
			case evt := <-r.chEvent:
				switch evt.Type {
				case RETClose:
					if isShuttingDown {
						break outerSelect
					}
					isShuttingDown = true
					for client := range r.clients {
						fmt.Printf("===== WsRouter telling client %v to shut down\n", client)
						client.BeginShutDown()
					}
				case RETBroadcast:
					if isShuttingDown {
						break outerSelect
					}
					for client := range r.clients {
						client.Send(evt.Payload)
					}
				default:
					panic(fmt.Errorf("WsRouter encountered unknown RET %d", evt.Type))
				}
			}
		}
	}()
}

func (r *WsRouter) Broadcast(msg []byte) {
	r.chEvent <- WsRouterEvent{Type: RETBroadcast, Payload: msg}
}

func (r *WsRouter) BeginShutdown() {
	fmt.Println("===== WsRouter sending RETClose (beginning shutdown)")
	r.chEvent <- WsRouterEvent{Type: RETClose, Payload: nil}
}

func (r *WsRouter) WaitForShutdown() {
	fmt.Println("===== WsRouter WaitForShutdown begin")
	r.wg.Wait()
	fmt.Println("===== WsRouter WaitForShutdown end")
}
