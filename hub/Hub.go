package hub

import (
	"time"

	"github.com/amirhossein-shakeri/go-sample/utils"
)

const (
	HUB_EVENT_LOG_CAPACITY = 1 << 10
	PING_REFRESH_INTERVAL  = 5 * time.Second
)

type Hub struct {
	Name       string                              `json:"name"`
	Clients    map[*Client]bool                    `json:"-"`
	Broadcast  chan []byte                         `json:"-"`
	Register   chan *Client                        `json:"-"`
	Unregister chan *Client                        `json:"-"`
	EventLog   utils.CircularList[Event[*Message]] `json:"events"`
}

/* All methods were removed except this one */

func (h *Hub) Run() {
	pingTicker := time.NewTicker(PING_REFRESH_INTERVAL)
	defer func() {
		pingTicker.Stop()
	}()

	for {
		select {
		case client, ok := <-h.Register:
			// register the client was here
			break
		case client := <-h.Unregister:
			// unregister the client was here
			break
		case <-pingTicker.C:
			// tidy up pings was here
			break
		case msg, ok := <-h.Broadcast:
			// broadcast logic was here
			break
		}
	}
}
