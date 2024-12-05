package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/rmatsuoka/sloppy-vr/server/internal/hatenaauth"
	"github.com/rmatsuoka/sloppy-vr/server/internal/hub"
	"github.com/rmatsuoka/sloppy-vr/server/internal/socksrv"
)

var (
	addr = flag.String("addr", "127.0.0.1:8001", "addr")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	mux := http.NewServeMux()

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	srv := socksrv.NewServer(&hub.Hub{
		Client:      client,
		ChannelName: "user.position",
	})
	srv.Install(func(pattern string, h http.Handler) {
		mux.Handle(pattern, hatenaauth.AuthHandler(h, func(w http.ResponseWriter, req *http.Request) {
			http.Error(w, "bad request", http.StatusBadRequest)
		}))
	})

	auth := hatenaauth.New()
	auth.Install(mux.Handle)

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		slog.Info("request",
			"method", req.Method,
			"url", req.URL.String(),
		)
		mux.ServeHTTP(w, req)
	})

	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal(err)
	}
}
