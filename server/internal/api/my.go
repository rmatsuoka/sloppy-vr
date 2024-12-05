package api

import (
	"encoding/json"
	"net/http"

	"github.com/rmatsuoka/sloppy-vr/server/internal/hatenaauth"
)

func Install(handle func(string, http.Handler)) {
	handle("GET /my", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		user, ok := hatenaauth.MyFromContext(req.Context())
		if !ok {
			return
		}
		buf, _ := json.Marshal(user)
		w.Write(buf)
	}))
}
