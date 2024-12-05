package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"

	"github.com/redis/go-redis/v9"
	sloppyvr "github.com/rmatsuoka/sloppy-vr"
	"github.com/rmatsuoka/sloppy-vr/server/internal/api"
	"github.com/rmatsuoka/sloppy-vr/server/internal/hatenaauth"
	"github.com/rmatsuoka/sloppy-vr/server/internal/hub"
	"github.com/rmatsuoka/sloppy-vr/server/internal/socksrv"
)

var (
	addr     = flag.String("addr", "127.0.0.1:8001", "addr")
	certFile = flag.String("cert", "", "TLS certFile")
	keyFile  = flag.String("key", "", "TLS keyFile")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	enableTLS := *certFile != "" || *keyFile != ""

	if enableTLS {
		if *certFile == "" {
			log.Fatal("no certFile")
		}
		if *keyFile == "" {
			log.Fatal("no keyFile")
		}
	}

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

	api.Install(func(pattern string, h http.Handler) {
		mux.Handle(pattern, hatenaauth.AuthHandler(h, func(w http.ResponseWriter, req *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		}))
	})

	mux.Handle("GET /styles/", http.FileServerFS(sloppyvr.FS))
	mux.Handle("GET /scripts/", http.FileServerFS(sloppyvr.FS))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, sloppyvr.FS, "index.html")
	})
	mux.HandleFunc("GET /vr", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, sloppyvr.FS, "vr.html")
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		slog.Info("request",
			"method", req.Method,
			"url", req.URL.String(),
		)
		mux.ServeHTTP(w, req)
	})

	var err error
	if enableTLS {
		err = http.ListenAndServeTLS(*addr, *certFile, *keyFile, handler)
	} else {
		err = http.ListenAndServe(*addr, handler)
	}
	if err != nil {
		log.Fatal(err)
	}
}
