package dog

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Second

	// Send ping interval, must be less then pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 32 * 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// CET is Client Event Type
type CET uint8

const (
	CETClose CET = iota
	CETSend
)

type WsClientEvent struct {
	Type    CET
	Payload []byte
}

// WsClient holds the open connection to individual websocket clients.
type WsClient struct {
	conn    *websocket.Conn
	owner   *WsRouter
	send    chan WsClientEvent
	wgRead  sync.WaitGroup
	wgWrite sync.WaitGroup
}

func CreateWsClient(conn *websocket.Conn, owner *WsRouter) *WsClient {
	c := &WsClient{
		conn:  conn,
		owner: owner,
		send:  make(chan WsClientEvent),
	}

	c.wgWrite.Add(1)
	go func() {
		defer c.wgWrite.Done()
		err := c.writePump()
		if err != nil {
			c.owner.HandleClientError(c, err)
		}
	}()

	c.wgRead.Add(1)
	go func() {
		defer c.wgRead.Done()
		err := c.readPump()
		if err != nil {
			c.owner.HandleClientError(c, err)
		}
	}()

	owner.register <- c
	return c
}

// Send is a thread-safe call to send a message to this client.
// Calling Send after ShutDown will panic.
func (c *WsClient) Send(msg []byte) {
	c.send <- WsClientEvent{Type: CETSend, Payload: msg}
}

// Close breaks the connection, waits for everything to halt, then updates the owner.
func (c *WsClient) BeginShutDown() {
	fmt.Println("----- WsClient sending CETClose (begin shut down)")
	c.send <- WsClientEvent{Type: CETClose, Payload: nil}
}

// Close breaks the connection, waits for everything to halt, then updates the owner.
func (c *WsClient) WaitForShutDown() {
	fmt.Println("----- WsClient.WatiForShutdown waiting for write pump")
	c.wgWrite.Wait()
	fmt.Println("----- WsClient.WatiForShutdown done")
}

func (c *WsClient) readPump() error {
	defer func() {
		fmt.Println("----- client read pump shutting down")
		c.BeginShutDown()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Start endless read loop, waiting for messages from client
	for {
		mtype, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return nil
			}
			return fmt.Errorf("read message: %w", err)
		}

		c.owner.HandleClientMessage(c, mtype, msg)
	}
}

func (c *WsClient) writePump() error {
	pingTimer := time.NewTicker(pingPeriod)
	defer func() {
		fmt.Println("----- client write pump shutting down")
		pingTimer.Stop()
		c.conn.Close()
		c.wgRead.Wait()
		c.owner.unregister <- c
	}()

	for {
		select {
		case <-pingTimer.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return fmt.Errorf("error sending websocket ping: %w", err)
			}
		case evt := <-c.send:
			switch evt.Type {
			case CETClose:
				fmt.Println("----- WsClient got CETClose")
				return nil
			case CETSend:
				// continue out of this switch
			default:
				panic(fmt.Errorf("WsClient encountered unknown CET %d", evt.Type))
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return fmt.Errorf("error getting next websocket writer: %w", err)
			}

			if _, err := w.Write(evt.Payload); err != nil {
				return fmt.Errorf("error sending websocket message: %w", err)
			}

			// n := len(c.send)
			// for i := 0; i < n; i++ {
			// 	w.Write([]byte{'\n'})
			// 	w.Write(<-c.send)
			// }

			if err := w.Close(); err != nil {
				return fmt.Errorf("error closing websocket writer: %w", err)
			}
		}
	}
}
