package web

import (
	"encoding/json"
	"net/http"
)

func (srv *Server) handleToggleMeal(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MemberID string `json:"member_id"`
		Date     string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.MemberID == "" || body.Date == "" {
		jsonErr(w, "Missing fields", 400)
		return
	}
	sets, _ := srv.store.GetSettings()
	if !canModify(body.Date, sets) {
		jsonErr(w, "Deadline passed — sign-up closed for this date", 403)
		return
	}
	checked, dayTotal, err := srv.store.ToggleMeal(body.Date, body.MemberID)
	if err != nil {
		jsonErr(w, "Update failed", 500)
		return
	}
	jsonOK(w, map[string]any{"checked": checked, "day_total": dayTotal})
}
