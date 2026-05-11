package web

import (
	"net/http"
	"sort"
	"time"
)

type HistoryEntry struct {
	Date  string
	Names []string
	Count int
}

type HistoryData struct {
	CurrentPage string
	Entries     []HistoryEntry
}

func (srv *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	today := time.Now().Format("2006-01-02")

	members, _ := srv.store.GetMembers()
	memberMap := map[string]string{}
	for _, m := range members {
		memberMap[m.ID] = m.Name
	}

	pastMeals, _ := srv.store.GetPastMeals(today)

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

	srv.render(w, "history.html", HistoryData{CurrentPage: "history", Entries: entries})
}
