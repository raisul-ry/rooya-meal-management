package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ── Global state ──────────────────────────────────────

var (
	store            Store
	baseDir          string
	menuFile         string
	settingsPassword string
)

func init() {
	baseDir, _ = os.Getwd()
	menuFile = filepath.Join(baseDir, "static", "menu.pdf")
	settingsPassword = os.Getenv("SETTINGS_PASSWORD")
}

// ── Store initialisation ──────────────────────────────

func initStore() Store {
	if connStr := os.Getenv("DATABASE_URL"); connStr != "" {
		pg, err := NewPGStore(connStr)
		if err != nil {
			log.Fatalf("postgres: %v", err)
		}
		log.Println("storage: PostgreSQL")
		return pg
	}
	log.Println("storage: local JSON files")
	return NewJSONStore(
		filepath.Join(baseDir, "data", "meals.json"),
		filepath.Join(baseDir, "data", "settings.json"),
	)
}

// ── Utilities ─────────────────────────────────────────

func newUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func canModify(dateStr string, s Settings) bool {
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return false
	}
	deadline := time.Date(t.Year(), t.Month(), t.Day(),
		s.DeadlineHour, s.DeadlineMinute, 0, 0, time.Local).
		AddDate(0, 0, -1)
	return time.Now().Before(deadline)
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func removeStr(slice []string, s string) []string {
	out := make([]string, 0, len(slice))
	for _, v := range slice {
		if v != s {
			out = append(out, v)
		}
	}
	return out
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

// ── Templates ─────────────────────────────────────────

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

func render(w http.ResponseWriter, page string, data any) {
	layout := filepath.Join(baseDir, "templates", "layout.html")
	pg := filepath.Join(baseDir, "templates", page)
	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(layout, pg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

// ── View models ───────────────────────────────────────

type DayCol struct {
	Date        string
	Modifiable  bool
	Total       int
	IsYesterday bool
}

type MealCell struct {
	MemberID    string
	Date        string
	Checked     bool
	Modifiable  bool
	IsYesterday bool
}

type MemberRow struct {
	ID    string
	Name  string
	Cells []MealCell
	Total int
}

type IndexData struct {
	CurrentPage string
	Members     []MemberRow
	Dates       []DayCol
	Total       int
	HasMembers  bool
}

type HistoryEntry struct {
	Date  string
	Names []string
	Count int
}

type HistoryData struct {
	CurrentPage string
	Entries     []HistoryEntry
}

type MenuData struct {
	CurrentPage string
	HasPDF      bool
}

type SettingsData struct {
	CurrentPage string
	Settings    Settings
	Hours       []int
	Minutes     []int
}

// ── Route handlers ────────────────────────────────────

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	sets, err := store.GetSettings()
	if err != nil {
		http.Error(w, "settings unavailable", 500)
		return
	}
	members, err := store.GetMembers()
	if err != nil {
		http.Error(w, "members unavailable", 500)
		return
	}

	today := time.Now().In(time.Local)
	dates := make([]DayCol, 7)
	dateStrs := make([]string, 7)
	for i := range dates {
		d := today.AddDate(0, 0, i-1).Format("2006-01-02")
		dates[i] = DayCol{Date: d, Modifiable: canModify(d, sets), IsYesterday: i == 0}
		dateStrs[i] = d
	}

	meals, err := store.GetMeals(dateStrs)
	if err != nil {
		http.Error(w, "meals unavailable", 500)
		return
	}

	rows := make([]MemberRow, len(members))
	for i, m := range members {
		cells := make([]MealCell, len(dates))
		memberTotal := 0
		for j, dc := range dates {
			chk := contains(meals[dc.Date], m.ID)
			if chk {
				memberTotal++
			}
			cells[j] = MealCell{
				MemberID:    m.ID,
				Date:        dc.Date,
				Checked:     chk,
				Modifiable:  dc.Modifiable,
				IsYesterday: dc.IsYesterday,
			}
		}
		rows[i] = MemberRow{ID: m.ID, Name: m.Name, Cells: cells, Total: memberTotal}
	}

	grand := 0
	for i, dc := range dates {
		t := len(meals[dc.Date])
		dates[i].Total = t
		grand += t
	}

	render(w, "index.html", IndexData{
		CurrentPage: "index",
		Members:     rows,
		Dates:       dates,
		Total:       grand,
		HasMembers:  len(rows) > 0,
	})
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	today := time.Now().Format("2006-01-02")

	members, _ := store.GetMembers()
	memberMap := map[string]string{}
	for _, m := range members {
		memberMap[m.ID] = m.Name
	}

	pastMeals, _ := store.GetPastMeals(today)

	var pastDates []string
	for date := range pastMeals {
		pastDates = append(pastDates, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(pastDates)))

	entries := make([]HistoryEntry, 0, len(pastDates))
	for _, date := range pastDates {
		mids := pastMeals[date]
		names := make([]string, 0, len(mids))
		for _, mid := range mids {
			n := memberMap[mid]
			if n == "" {
				n = "(removed)"
			}
			names = append(names, n)
		}
		entries = append(entries, HistoryEntry{Date: date, Names: names, Count: len(names)})
	}

	render(w, "history.html", HistoryData{CurrentPage: "history", Entries: entries})
}

func handleMenu(w http.ResponseWriter, r *http.Request) {
	_, err := os.Stat(menuFile)
	render(w, "menu.html", MenuData{CurrentPage: "menu", HasPDF: err == nil})
}

func handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if settingsPassword != "" && r.FormValue("password") != settingsPassword {
			jsonErr(w, "Wrong password", 403)
			return
		}
		hour, _ := strconv.Atoi(r.FormValue("deadline_hour"))
		minute, _ := strconv.Atoi(r.FormValue("deadline_minute"))
		store.SaveSettings(Settings{DeadlineHour: hour, DeadlineMinute: minute})
		jsonOK(w, map[string]bool{"success": true})
		return
	}
	hours := make([]int, 24)
	for i := range hours {
		hours[i] = i
	}
	s, _ := store.GetSettings()
	render(w, "settings.html", SettingsData{
		CurrentPage: "settings",
		Settings:    s,
		Hours:       hours,
		Minutes:     []int{0, 15, 30, 45},
	})
}

// ── API handlers ──────────────────────────────────────

func handleMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	var body struct {
		Names    []string `json:"names"`
		Password string   `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "Invalid JSON", 400)
		return
	}
	if settingsPassword != "" && body.Password != settingsPassword {
		jsonErr(w, "Wrong password", 403)
		return
	}
	added, skipped := 0, 0
	for _, raw := range body.Names {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		exists, err := store.MemberNameExists(name)
		if err != nil {
			jsonErr(w, "DB error", 500)
			return
		}
		if exists {
			skipped++
			continue
		}
		if err := store.AddMember(Member{ID: newUUID(), Name: name}); err != nil {
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

func handleMember(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/members/")
	if id == "" {
		jsonErr(w, "ID required", 400)
		return
	}
	if r.Method != http.MethodDelete {
		http.NotFound(w, r)
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if settingsPassword != "" && body.Password != settingsPassword {
		jsonErr(w, "Wrong password", 403)
		return
	}
	if err := store.DeleteMember(id); err != nil {
		jsonErr(w, "Delete failed", 500)
		return
	}
	jsonOK(w, map[string]bool{"success": true})
}

func handleToggleMeal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	var body struct {
		MemberID string `json:"member_id"`
		Date     string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.MemberID == "" || body.Date == "" {
		jsonErr(w, "Missing fields", 400)
		return
	}
	sets, _ := store.GetSettings()
	if !canModify(body.Date, sets) {
		jsonErr(w, "Deadline passed — sign-up closed for this date", 403)
		return
	}
	checked, dayTotal, err := store.ToggleMeal(body.Date, body.MemberID)
	if err != nil {
		jsonErr(w, "Update failed", 500)
		return
	}
	jsonOK(w, map[string]any{"checked": checked, "day_total": dayTotal})
}

func handleMenuUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		jsonErr(w, "Request too large", 400)
		return
	}
	file, header, err := r.FormFile("pdf")
	if err != nil {
		jsonErr(w, "No file uploaded", 400)
		return
	}
	defer file.Close()
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		jsonErr(w, "Only PDF files allowed", 400)
		return
	}
	dst, err := os.Create(menuFile)
	if err != nil {
		jsonErr(w, "Failed to save file", 500)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)
	jsonOK(w, map[string]bool{"success": true})
}

// ── Main ──────────────────────────────────────────────

func main() {
	if tz := os.Getenv("TZ"); tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			time.Local = loc
		}
	}
	store = initStore()

	mux := http.NewServeMux()

	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/history", handleHistory)
	mux.HandleFunc("/menu", handleMenu)
	mux.HandleFunc("/settings", handleSettings)

	mux.HandleFunc("/api/members", handleMembers)
	mux.HandleFunc("/api/members/", handleMember)
	mux.HandleFunc("/api/meals/toggle", handleToggleMeal)
	mux.HandleFunc("/api/menu/upload", handleMenuUpload)

	mux.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir(filepath.Join(baseDir, "static")))))

	log.Println("Meal Manager → http://localhost:5050")
	log.Fatal(http.ListenAndServe(":5050", mux))
}
