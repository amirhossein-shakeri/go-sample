package hub

type MessageType int

const (
	MTUndefined MessageType = iota
	MTStateUpdate
	MTControlCommand
	MTMessage
)

const (
	StatusOK        = 4200 + iota
	StatusCreated   // 4201: New Item was added or created
	_               // Skip 4202
	_               // Skip 4203
	StatusNoContent // 4204: Nothing to return like a delete
)

const (
	StatusBadRequest   = 4400 + iota
	StatusUnauthorized // UnAuthenticated. Similar to HTTP 401
	_                  // todo
	StatusForbidden    // UnAuthorized. Similar to HTTP 403
	StatusNotFound     // Unknown Action Not Found.
)

/* Define message schema */
type Message struct {
	Signal  string                 `json:"signal,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}
