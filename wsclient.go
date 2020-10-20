package dog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
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

// WsClient holds the open connection to individual websocket clients.
// It manages reading, writing, and lifetime.
type WsClient struct {
	conn    *websocket.Conn
	owner   *WsRouter
	chSend  chan wsClientEvent
	chDead  chan struct{}
	wgRead  sync.WaitGroup
	wgWrite sync.WaitGroup
}

func CreateWsClient(conn *websocket.Conn, owner *WsRouter) *WsClient {
	c := &WsClient{
		conn:   conn,
		owner:  owner,
		chSend: make(chan wsClientEvent),
		chDead: make(chan struct{}),
	}

	c.wgWrite.Add(1)
	go func() {
		defer c.wgWrite.Done()
		err := c.writePump()
		if err != nil {
			c.owner.handleClientError(c, err)
		}
	}()

	c.wgRead.Add(1)
	go func() {
		defer c.wgRead.Done()
		err := c.readPump()
		if err != nil {
			c.owner.handleClientError(c, err)
		}
	}()

	// build initial message
	var buf bytes.Buffer
	buf.Grow(256)
	buf.WriteString(`{"index_hash":"`)
	buf.WriteString(owner.indexHash)
	buf.WriteString(`"}`)
	err := c.sendImpl(cstInit, buf.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	return c
}

// Send is a thread-safe call to send a message to this client.
// Send will nop if called after client begins shutting down.
func (c *WsClient) Send(msg json.RawMessage) {
	err := c.sendImpl(cstLog, msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
}

// BeginShutDown requests that this client shut itself down.
// BeginShutDown will nop if called after client begins shutting down.
func (c *WsClient) BeginShutDown() {
	c.queueEvent(cetClose, nil)
}

// Close breaks the connection, waits for everything to halt, then updates the owner.
func (c *WsClient) WaitForShutDown() {
	// the write pump waits for the read pump to end before ending itself, so just wait for that.
	c.wgWrite.Wait()
}

type clientSendType uint8

// clientSendType needs to stay in sync with the javascript message receiver
const (
	cstInit clientSendType = 0
	cstLog  clientSendType = 1
)

type clientSendPayload struct {
	Type clientSendType  `json:"t"`
	Msg  json.RawMessage `json:"m"`
}

func (c *WsClient) sendImpl(cst clientSendType, msg json.RawMessage) error {
	b, err := json.Marshal(&clientSendPayload{Type: cst, Msg: msg})
	if err != nil {
		return fmt.Errorf("error building message to web client: %w", err)
	}
	c.queueEvent(cetSend, b)
	return nil
}

type clientEventType uint8

const (
	cetClose clientEventType = iota
	cetSend
)

type wsClientEvent struct {
	cet     clientEventType
	payload json.RawMessage
}

func (c *WsClient) queueEvent(cet clientEventType, payload json.RawMessage) {
	select {
	case <-c.chDead:
		// If chDone is closed, then chSend will never be serviced.
	case c.chSend <- wsClientEvent{cet: cet, payload: payload}:
	}
}

func (c *WsClient) readPump() error {
	defer func() {
		// If this pump is ending due to a connection error, then trigger a shutdown.
		// If the shutdown triggered the end of the read pump, this will nop.
		// Either way, the read pump has to go down before the write pump can end.
		c.BeginShutDown()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		mtype, msg, err := c.conn.ReadMessage()
		if err != nil {
			ignoreErr := websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway)
			select {
			case <-c.chDead:
				ignoreErr = true
			default:
			}
			if ignoreErr {
				return nil
			}
			return fmt.Errorf("read message: %w", err)
		}

		if mtype == websocket.TextMessage {
			c.owner.handleClientMessage(c, msg)
		}
	}
}

func (c *WsClient) writePump() error {
	pingTimer := time.NewTicker(pingPeriod)
	defer func() {
		pingTimer.Stop()
		// mark this client as dead, so any requests or connection errors will be ignored during shutdown
		close(c.chDead)
		// if this connection isn't already dead, then kill it
		c.conn.Close()
		// read pump should die now that the connection is dead, so wait for it
		c.wgRead.Wait()
		// finally tell the owner we no longer exist
		c.owner.handleClientShutdown(c)
	}()

	for {
		select {
		case <-pingTimer.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return fmt.Errorf("error sending websocket ping: %w", err)
			}
		case evt := <-c.chSend:
			switch evt.cet {
			case cetClose:
				// write pump needs to cleanly shut itself down
				return nil
			case cetSend:
				// continue out of this switch
			default:
				return fmt.Errorf("WsClient encountered unknown cet %d", evt.cet)
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return fmt.Errorf("error getting next websocket writer: %w", err)
			}

			if _, err := w.Write(evt.payload); err != nil {
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
