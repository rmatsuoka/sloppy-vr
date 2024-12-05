package socksrv

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rmatsuoka/sloppy-vr/server/internal/hub"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 1024
)

type Server struct {
	upgrader *websocket.Upgrader
	hub      *hub.Hub
}

func NewServer(hub *hub.Hub) *Server {
	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	return &Server{
		upgrader: upgrader,
		hub:      hub,
	}
}

func (s *Server) Install(handle func(string, http.Handler)) {
	handleFunc := func(pattern string, h http.HandlerFunc) {
		handle(pattern, h)
	}

	handleFunc("/socketserver", s.PubSub)
}

func (s *Server) PubSub(w http.ResponseWriter, req *http.Request) {
	conn, err := s.upgrader.Upgrade(w, req, w.Header())
	if err != nil {
		slog.Error("upgrader.Upgrade", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	ch := make(chan struct{})
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	subscriber, err := s.hub.Subscribe(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to subscribe", "error", err)
		conn.Close()
		return
	}
	defer subscriber.Close()

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
			ch <- struct{}{}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case mesg := <-subscriber.Channel():
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				slog.Info("subscribe mesg", "message", mesg)
				err := conn.WriteMessage(websocket.TextMessage, []byte(mesg))
				if err != nil {
					slog.Error("error on conn.WriteMessage",
						"remoteAddr", conn.RemoteAddr(),
						"error", err,
					)
					return
				}
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					slog.ErrorContext(ctx, "writeMessage ping", "error", err)
					return
				}
			}
		}
	}()

	go func() {
		defer func() { ch <- struct{}{} }()
		conn.SetReadLimit(maxMessageSize)
		conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_, message, err := conn.ReadMessage()
			if err != nil {
				slog.Error("error on conn.ReadMessage",
					"remoteAddr", conn.RemoteAddr(),
					"error", err,
				)
				return
			}
			// slog.Info("websock message", "message", message)
			err = s.hub.Publish(ctx, string(message))
			if err != nil {
				slog.ErrorContext(ctx, "error on s.hub.Publish", "error", err)
				return
			}
		}
	}()

	// reader or writer has finished.
	<-ch
	cancel()

	<-ch // wait for canceling writer or reader.
}
