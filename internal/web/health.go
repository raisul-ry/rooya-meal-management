package web

import (
	"net/http"
	"time"
)

func (srv *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonOK(w, map[string]string{"status": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
}
