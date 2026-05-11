package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/raisul-ry/rooya-meal-management/internal/store"
)

func (srv *Server) handleMembers(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Names    []string `json:"names"`
		Password string   `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "Invalid JSON", 400)
		return
	}
	if !srv.checkPassword(body.Password) {
		jsonErr(w, "Wrong password", 403)
		return
	}
	added, skipped := 0, 0
	for _, raw := range body.Names {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		exists, err := srv.store.MemberNameExists(name)
		if err != nil {
			jsonErr(w, "DB error", 500)
			return
		}
		if exists {
			skipped++
			continue
		}
		if err := srv.store.AddMember(store.Member{ID: newUUID(), Name: name}); err != nil {
			jsonErr(w, "Save failed", 500)
			return
		}
		added++
	}
	if added == 0 && skipped == 0 {
		jsonErr(w, "No names provided", 400)
		return
	}
	jsonOK(w, map[string]int{"added": added, "skipped": skipped})
}

func (srv *Server) handleMember(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonErr(w, "ID required", 400)
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if !srv.checkPassword(body.Password) {
		jsonErr(w, "Wrong password", 403)
		return
	}
	if err := srv.store.DeleteMember(id); err != nil {
		jsonErr(w, "Delete failed", 500)
		return
	}
	jsonOK(w, map[string]bool{"success": true})
}
