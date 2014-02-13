package websockets

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
)

type WebsocketsHandler struct {
}

func New() (*WebsocketsHandler, error) {
	return &WebsocketsHandler{}, nil
}

func (websocketHandler *WebsocketsHandler) WireWebsocketSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/connect", websocketHandler.upgrade).Methods("GET")
}

func (websocketHandler *WebsocketsHandler) upgrade(writer http.ResponseWriter, request *http.Request) {
	socket, err := websocket.Upgrade(writer, request, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(writer, "Not a websocket handshake", http.StatusBadRequest)
		return
	} else if err != nil {
		panic(err)
	}

	if err = socket.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
		panic(err)
	}
}
