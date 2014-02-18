package websockets

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

type WebsocketsManager struct {
	websocketConns map[string]*websocket.Conn
}

func NewManager() (*WebsocketsManager, error) {
	websocketManager := WebsocketsManager{make(map[string]*websocket.Conn)}
	go websocketManager.startHeartbeat()
	return &websocketManager, nil
}

func (websocketsManager *WebsocketsManager) GetId(userId, connectionId uint64) string {
	return fmt.Sprintf("%d-%d", userId, connectionId)
}

func (websocketsManager *WebsocketsManager) Add(id string, websocketConn *websocket.Conn) error {
	if _, ok := websocketsManager.websocketConns[id]; ok {
		return fmt.Errorf("Websocket already exists with id %s", id)
	}

	websocketsManager.websocketConns[id] = websocketConn
	return nil
}

func (websocketsManager *WebsocketsManager) SendJson(id string, message interface{}) error {
	websocketConn, ok := websocketsManager.websocketConns[id]
	if !ok {
		return fmt.Errorf("Unable to find websocket with id: %d", id)
	}

	if err := websocketConn.WriteJSON(message); err != nil {
		delete(websocketsManager.websocketConns, id)
		return err
	}
	return nil
}

func (websocketsManager *WebsocketsManager) startHeartbeat() {
	timer := time.Tick(45 * time.Second)
	for _ = range timer {
		websocketsManager.performHeartbeat()
	}
}

func (websocketsManager *WebsocketsManager) performHeartbeat() {
	for id, websocketConn := range websocketsManager.websocketConns {
		err := websocketConn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(10*time.Second))
		if err != nil {
			delete(websocketsManager.websocketConns, id)
		}
	}
}
