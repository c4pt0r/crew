package main

import (
	"net/http"
	"strings"

	"github.com/c4pt0r/log"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.E(err)
		return
	}
	defer conn.Close()

	// Handle WebSocket messages here
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.E(err)
			return
		}
		// You can process WebSocket messages here and send responses back to the client
		// For example:
		response := []byte("WebSocket Server: Received message: " + string(p))
		conn.WriteMessage(messageType, response)

	}
}

type adminCmdhandlerFunc func(*adminConn, []byte)

type adminConn struct {
	Conn   *websocket.Conn
	token  string
	authed bool

	// handlers
	handlers map[string]adminCmdhandlerFunc
}

func (ac *adminConn) send(msg []byte) {
	ac.Conn.WriteMessage(websocket.TextMessage, msg)
}

func (ac *adminConn) close() {
	ac.Conn.Close()
}

func (ac *adminConn) read() ([]byte, error) {
	_, msg, err := ac.Conn.ReadMessage()
	return msg, err
}

func (ac *adminConn) welcome() {
	ac.send([]byte("Welcome to the admin console"))
}

func (ac *adminConn) handle(msg []byte) {

}

func (ac *adminConn) handleAuth(msg []byte) {
	if ac.authed {
		ac.send([]byte("Already authed"))
		return
	}
	line := string(msg)
	fields := strings.Fields(line)
	if len(fields) != 2 {
		ac.send([]byte("Invalid auth command"))
		return
	}
	if fields[0] != "auth" {
		ac.send([]byte("Invalid auth command"))
		return
	}
}
