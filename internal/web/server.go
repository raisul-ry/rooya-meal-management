package web

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/raisul-ry/rooya-meal-management/internal/store"
)

// Server holds shared dependencies for all HTTP handlers.
type Server struct {
	store    store.Store
	password string
	baseDir  string
	menuFile string
}

func NewServer(s store.Store, password, baseDir string) *Server {
	return &Server{
		store:    s,
		password: password,
		baseDir:  baseDir,
		menuFile: filepath.Join(baseDir, "static", "menu.pdf"),
	}
}

func (srv *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", srv.handleIndex)
	mux.HandleFunc("GET /history", srv.handleHistory)
	mux.HandleFunc("GET /menu", srv.handleMenu)
	mux.HandleFunc("GET /settings", srv.handleSettings)
	mux.HandleFunc("POST /settings", srv.handleSettings)
	mux.HandleFunc("POST /api/members", srv.handleMembers)
	mux.HandleFunc("DELETE /api/members/{id}", srv.handleMember)
	mux.HandleFunc("POST /api/meals/toggle", srv.handleToggleMeal)
	mux.HandleFunc("POST /api/menu/upload", srv.handleMenuUpload)
	mux.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir(filepath.Join(srv.baseDir, "static")))))

	return mux
}

var funcMap = template.FuncMap{
	"fmtDay": func(date string) string {
		t, _ := time.ParseInLocation("2006-01-02", date, time.Local)
		return t.Format("Mon")
	},
	"fmtDate": func(date string) string {
		t, _ := time.ParseInLocation("2006-01-02", date, time.Local)
		return t.Format("Jan 02")
	},
	"fmtFull": func(date string) string {
		t, _ := time.ParseInLocation("2006-01-02", date, time.Local)
		return t.Format("Monday, January 02, 2006")
	},
	"plural": func(n int, word string) string {
		if n == 1 {
			return word
		}
		return word + "s"
	},
	"add1": func(i int) int { return i + 1 },
}

func (srv *Server) render(w http.ResponseWriter, page string, data any) {
	layout := filepath.Join(srv.baseDir, "templates", "layout.html")
	pg := filepath.Join(srv.baseDir, "templates", page)
	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(layout, pg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (srv *Server) checkPassword(pw string) bool {
	return srv.password == "" || pw == srv.password
}

func canModify(dateStr string, s store.Settings) bool {
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return false
	}
	deadline := time.Date(t.Year(), t.Month(), t.Day(),
		s.DeadlineHour, s.DeadlineMinute, 0, 0, time.Local).
		AddDate(0, 0, -1)
	return time.Now().Before(deadline)
}

func newUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func jsonErr(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
