package main

import (
	"bytes"
	"context"
	"errors"
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

	ac := &adminConn{
		wsConn: conn,
		token:  "hello",
		authed: false,
	}
	defer ac.close()
	ac.welcome()

	if !ac.auth() {
		ac.send("need auth")
		return
	} else {
		ac.send("OK")
	}

	ac.serve()
}

type adminConn struct {
	wsConn *websocket.Conn
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
	ac.send("Welcome")
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

type AdminConnContext context.Context

func NewAdminConnContext(ctx context.Context, ac *adminConn) AdminConnContext {
	return AdminConnContext(context.WithValue(ctx, "adminConn", ac))
}

func GetAdminConnContext(ctx context.Context) *adminConn {
	return ctx.Value("adminConn").(*adminConn)
}

func (n *node) walk(fn func(n *node)) {
	fn(n)
	subnodes, err := n.getSubNodes()
	if err != nil {
		return
	}
	for _, c := range subnodes {
		c.walk(fn)
	}
}

var (
	cmdKeys = []string{
		"quit",
		"echo",
		"help",
	}
	cmds = map[string]func(ctx context.Context, cmd string, params []string) error{
		"quit": func(ctx context.Context, cmd string, params []string) error {
			ac := GetAdminConnContext(ctx)
			ac.send("bye")
			ac.close()
			return errors.New("quit")
		},
		"echo": func(ctx context.Context, cmd string, params []string) error {
			ac := GetAdminConnContext(ctx)
			ac.send(strings.Join(params, " "))
			return nil
		},
		"buildindex": func(ctx context.Context, cmd string, params []string) error {
			ac := GetAdminConnContext(ctx)
			ac.send("buildindex")
			return nil
		},
		"list": func(ctx context.Context, cmd string, params []string) error {
			ac := GetAdminConnContext(ctx)
			root := getRootNode()
			buf := bytes.NewBuffer(nil)
			root.walk(func(n *node) {
				buf.WriteString(n.filepath)
				buf.WriteString("\n")
			})
			ac.send(buf.String())
			return nil
		},
		"buildindex": func(ctx context.Context, cmd string, params []string) error {
			// send ok
			ac := GetAdminConnContext(ctx)
			ac.send("OK")
			return nil
		},
		"help": func(ctx context.Context, cmd string, params []string) error {
			ac := GetAdminConnContext(ctx)
			ac.send(strings.Join(cmdKeys, " "))
			return nil
		},
	}
)

func (ac *adminConn) serve() error {
	for {
		cmd, params, err := ac.readCommand()
		if err != nil {
			return err
		}
		if fn, ok := cmds[cmd]; ok {
			ctx := NewAdminConnContext(context.Background(), ac)
			err := fn(ctx, cmd, params)
			if err != nil {
				return err
			}
		} else {
			ac.send("unknown command")
		}
	}
	return nil
}
