package socksrv

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Server struct {
	upgrader *websocket.Upgrader
}

func NewServer() *Server {
	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	return &Server{
		upgrader: upgrader,
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

	ch := make(chan struct{})
	canceled := make(chan struct{})

	go func() {
		defer func() { ch <- struct{}{} }()
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-canceled:
				slog.Info("writer is canceled")
				return
			case <-ticker.C:
				err := conn.WriteMessage(websocket.TextMessage, []byte("hello"))
				if err != nil {
					slog.Error("error on conn.WriteMessage",
						"remoteAddr", conn.RemoteAddr(),
						"error", err,
					)
					return
				}
			}
		}
	}()

	go func() {
		defer func() { ch <- struct{}{} }()
		for {
			select {
			case <-canceled:
				slog.Info("reader is canceled")
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
			slog.Info("conn.ReadMessage",
				"message", message,
				"remoteAddr", conn.RemoteAddr(),
			)
		}
	}()

	// reader or writer is finished.
	<-ch
	close(canceled)

	<-ch // wait for canceling writer or reader.
	conn.Close()
	slog.Info("conn.Close()")
}
