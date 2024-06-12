// I've added a 'Z' letter to the beginning of the function to move it down in the directory since the code has so much removals because of privacy stuff

package hub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/amirhossein-shakeri/go-sample/utils/log"
	"github.com/amirhossein-shakeri/websocket"
	"github.com/amirhossein-shakeri/websocket/wsjson"
	"golang.org/x/time/rate"
)

// Optimize WS server memory: https://www.freecodecamp.org/news/million-websockets-and-go-cc58418460bb

const (
	IdleTimeout       = 61 * time.Minute // If a client doesn't send or receive any message within the interval, it will be disconnected.
	ReadTimeout       = 61 * time.Minute // If a client doesn't send any message withing this interval, it will be disconnected.
	WriteTimeout      = 30 * time.Second // Time each write to the client has.
	PingTimeout       = 30 * time.Second // Time each client has to respond to ping message.
	PingInterval      = 20 * time.Second // Ping period. Server periodically pings clients each `PingInterval`. 60 * time.Minute
	MaxMessageSize    = 1024 * 128       // default: 128 KiB
	ReadLimitInterval = 1 * time.Second  // One server read every ...
	ReadLimitBurst    = 30               // Read Burst
)

type ClientType string

const (
	ClientTypeUndefined ClientType = ""
	ClientTypeDevice    ClientType = "Device"
	ClientTypeUser      ClientType = "User"
)

type Client struct {
	Conn      *websocket.Conn `json:"-"`    // websocket connection
	Hub       *Hub            `json:"-"`    // Address of the hub
	Send      chan *Message   `json:"-"`    // Buffered channel of outbound messages
	Type      ClientType      `json:"type"` // Device or User or anything else
	idleTimer *time.Timer     `json:"-"`    // idle timer used to kick idle clients after timeout
}

func NewClient( /* ... */ ) (*Client, error) {
	// ...
	return &Client{ /* ... */ }, nil
}

type ReadPumpHandler func(*Client, int, []byte, error) (bool, bool, error) // continue, break, error
type WritePumpHandler func(*Client, *Message, io.WriteCloser)

/*
writePump pumps messages from the hub to the websocket connection.
A goroutine running writePump is started for each connection. The
application ensures that there is at most one writer to a connection by
executing all writes from this goroutine.
*/
func (c *Client) WritePump() {
	ticker := time.NewTicker(PingInterval)
	defer func() {
		log.Debugf(
			"write pump is about to close the ws connection for client",
		)
		ticker.Stop()
		c.Hub.Unregister <- c
		c.Conn.Close(websocket.StatusNormalClosure, "write pump closed")
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.idleTimer.Reset(IdleTimeout)
			writeCtx, cancel := context.WithTimeout(context.TODO(), WriteTimeout)
			defer cancel()

			if !ok { // The hub closed the channel.
				c.Conn.Close(websocket.StatusNormalClosure, fmt.Sprintf("%s closed the channel", c.Hub.Name))
				log.Errorf("%s closed the channel. OK is `false`.", c.Hub.Name)
				return
			}
			if err := wsjson.Write(writeCtx, c.Conn, message); err != nil {
				log.Errorf("error writing wsjson: %v", err)
				go c.Hub.UseQuota(0, 0, 0, 1) // count x1 Failed
				return
			}
			go c.Hub.UseQuota(0, 1, 0, 0) // count x1 Sent

			if os.Getenv("DEBUG_SOCKET") == "true" && message.Signal != SignalRefreshLatencies {
				log.Debugf("...")
			}

		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(context.TODO(), PingInterval)
			defer cancel()
			pingStart := time.Now()
			if err := c.Conn.Ping(pingCtx); err != nil {
				log.Debugf("...")
				return // used to be break
			}
			c.idleTimer.Reset(IdleTimeout)
			c.Latency = time.Since(pingStart)
		}
	}
}

/*
readPump pumps messages from the websocket connection to the hub.

The application runs readPump in a per-connection goroutine. The application
ensures that there is at most one reader on a connection by executing all
reads from this goroutine.
*/
func (c *Client) ReadPump() {
	defer func() {
		log.Debugf("read pump is about to unregister the client")
		c.Hub.Unregister <- c
		c.Conn.Close(websocket.StatusNormalClosure, "read pump closed")
	}()

	// Max Size and Rate Limits
	c.Conn.SetReadLimit(MaxMessageSize)
	readLimiter := rate.NewLimiter(rate.Every(ReadLimitInterval), ReadLimitBurst)

	for {
		msg := Message{}
		// Creating a context with cancel so that we can handle the cancel with a timer and reset. Don't use context because of https://stackoverflow.com/questions/61455051/how-can-i-extend-the-timeout-of-a-context-in-go
		readContext, cancel := context.WithCancel(context.TODO())
		defer cancel()
		go func() {
			<-c.idleTimer.C
			cancel()
		}()

		// rate limiting stuff was here ...

		_, msgBytes, err := c.Conn.Read(readContext)

		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Debugf("read timed out: %v", err)
				return
			}
			closeStatus := websocket.CloseStatus(err)
			if closeStatus == websocket.StatusNormalClosure || closeStatus == websocket.StatusGoingAway {
				log.Debugf("client closed the connection. read pump is stopping ...")
				return
			}
			log.Infof("client disconnected: error reading message from client")
			go c.Hub.UseQuota(0, 0, 0, 1) // count x1 Failed
			return
		}

		readStart := time.Now()
		c.idleTimer.Reset(IdleTimeout) // Reset the idle timer

		// Parse JSON
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			log.Debugf("Error parsing JSON: %v, MSG: %v", err, string(msgBytes))
			c.Send <- MessageErrf("error parsing JSON: %v", err.Error())
			continue
		}

		// Use the quota
		if os.Getenv("DEBUG_SOCKET") == "true" {
			log.Debugf("[<-RSV]: 🟠 %s(%s) %v", c.Name(), c.ID(), msg)
		}
		go c.Hub.UseQuota(1, 0, 1, 0) // count x1 Used + x1 Received

		/* signal handling was here */

	}
}
