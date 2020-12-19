package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// Proxy connects to iguagile-engine via tcp on behalf of the client connecting via websocket.
type Proxy struct {
	Logger *log.Logger
	ug     *websocket.Upgrader
}

// New creates an instance of Proxy.
func New() *Proxy {
	return &Proxy{
		Logger: log.New(os.Stdout, "iguagile-ws-proxy", log.Lshortfile),
		ug: &websocket.Upgrader{
			CheckOrigin: func(*http.Request) bool { return true },
		},
	}
}

// Start starts an proxy server.
func (p *Proxy) Start(address string) error {
	return http.ListenAndServe(address, p)
}

// Server is iguagile-engine server.
type Server struct {
	Host string `json:"host"`
	Port int32  `json:"port"`
}

// Room has the information needed to connect to the room.
type Room struct {
	RoomId          int32             `json:"room_id"`
	Server          Server            `json:"server"`
	ApplicationName string            `json:"application_name"`
	Version         string            `json:"version"`
	Password        string            `json:"password,omitempty"`
	Token           string            `json:"token,omitempty"`
	Information     map[string]string `json:"information,omitempty"`
}

// ServeHTTP connects to iguagile-engine via tcp on behalf of the client connecting via websocket.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := p.ug.Upgrade(w, r, nil)
	if err != nil {
		p.Logger.Println(err)
		return
	}
	defer ws.Close()

	_, message, err := ws.ReadMessage()
	if err != nil {
		p.Logger.Println(err)
		return
	}

	var room Room
	if err := json.Unmarshal(message, &room); err != nil {
		p.Logger.Println(err)
		return
	}

	addr := fmt.Sprintf("%v:%v", room.Server.Host, room.Server.Port)
	tcp, err := dial(addr)
	if err != nil {
		p.Logger.Println(err)
		return
	}
	defer tcp.conn.Close()

	if err := tcp.verify(room); err != nil {
		p.Logger.Println(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := tcp.readStart(ctx, ws); err != nil {
			p.Logger.Println(err)
			cancel()
		}
	}()

	go func() {
		if err := tcp.writeStart(ctx, ws); err != nil {
			p.Logger.Println(err)
			cancel()
		}
	}()

	t := time.Tick(time.Second * 10)
	for {
		<-t
		if err := ws.WriteMessage(websocket.PongMessage, nil); err != nil {
			p.Logger.Println(err)
			cancel()
			return
		}
	}
}
