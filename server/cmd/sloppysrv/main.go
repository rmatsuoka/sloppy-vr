package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"

	"github.com/rmatsuoka/sloppy-vr/server/internal/socksrv"
)

var (
	addr = flag.String("addr", "127.0.0.1:8001", "addr")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	mux := http.NewServeMux()

	srv := socksrv.NewServer()
	srv.Install(mux.Handle)

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
