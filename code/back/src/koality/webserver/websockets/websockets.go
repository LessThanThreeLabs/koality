package websockets

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"koality/webserver/middleware"
	"net/http"
	"strconv"
)

type WebsocketsHandler struct {
	manager *WebsocketsManager
}

func New(manager *WebsocketsManager) (*WebsocketsHandler, error) {
	return &WebsocketsHandler{manager}, nil
}

func (websocketsHandler *WebsocketsHandler) WireWebsocketSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/connect",
		middleware.IsLoggedInWrapper(websocketsHandler.upgrade)).
		Methods("GET")
}

func (websocketsHandler *WebsocketsHandler) upgrade(writer http.ResponseWriter, request *http.Request) {
	userId := context.Get(request, "userId").(uint64)

	queryValues := request.URL.Query()
	connectionIdString := queryValues.Get("connectionId")
	connectionId, err := strconv.ParseUint(connectionIdString, 10, 64)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Unable to parse connectionId: %v", err), http.StatusInternalServerError)
		return
	}

	websocketConn, err := websocket.Upgrade(writer, request, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(writer, "Not a websocket handshake", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(writer, fmt.Sprintf("Error when upgrading connection: %v", err), http.StatusInternalServerError)
		return
	}

	websocketId := websocketsHandler.manager.GetId(userId, connectionId)
	websocketsHandler.manager.Add(websocketId, websocketConn)
}
