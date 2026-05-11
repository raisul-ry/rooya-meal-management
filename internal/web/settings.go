package web

import (
	"net/http"
	"strconv"

	"github.com/raisul-ry/rooya-meal-management/internal/store"
)

type SettingsData struct {
	CurrentPage string
	Settings    store.Settings
	Hours       []int
	Minutes     []int
}

func (srv *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if !srv.checkPassword(r.FormValue("password")) {
			jsonErr(w, "Wrong password", 403)
			return
		}
		hour, _ := strconv.Atoi(r.FormValue("deadline_hour"))
		minute, _ := strconv.Atoi(r.FormValue("deadline_minute"))
		srv.store.SaveSettings(store.Settings{DeadlineHour: hour, DeadlineMinute: minute})
		jsonOK(w, map[string]bool{"success": true})
		return
	}
	hours := make([]int, 24)
	for i := range hours {
		hours[i] = i
	}
	s, _ := srv.store.GetSettings()
	srv.render(w, "settings.html", SettingsData{
		CurrentPage: "settings",
		Settings:    s,
		Hours:       hours,
		Minutes:     []int{0, 15, 30, 45},
	})
}
