package main

import (
	"net/http"
	"strings"
	"errors"

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

	ac := &adminConn{
		wsConn: conn,
		token: "hello",
		authed: false,
	}
	defer ac.close()
	ac.welcome()

	if !ac.auth() {
		ac.send("need auth")
		return
	}

	ac.serve()
}

type adminConn struct {
	wsConn   *websocket.Conn
	token  string
	authed bool
}

func (ac *adminConn) send(msg string) {
	ac.wsConn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (ac *adminConn) close() {
	ac.wsConn.Close()
}


func (ac *adminConn) readCommand() (string, []string, error) {
	_, msg, err := ac.wsConn.ReadMessage()
	if err != nil {
		return "", nil, err
	}
	fields := strings.Fields(string(msg))
	if len(fields) == 0 {
		return "", nil, errors.New("invalid command")
	}
	cmd := strings.ToLower(fields[0])
	return cmd, fields[1:], nil
}

func (ac *adminConn) welcome() {
	ac.send("Welcome to the admin console")
}

func (ac *adminConn) auth() bool {
	cmd, params, err := ac.readCommand()
	if err != nil {
		log.E(err)
		return false
	}
	if cmd != "auth" || len(params) == 0 {
		return false
	}

	if ac.token != params[0] {
		return false
	}
	return true
}

func (ac *adminConn) serve() error {
	for {
		cmd, params, err := ac.readCommand()
		if err != nil {
			return err
		}
		switch cmd {
		case "quit":
			ac.send("bye!")
			return nil
		case "echo":
			ac.send(strings.Join(params, " "))
			return nil
		}
	}
	return nil
}

