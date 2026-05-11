package web

import (
	"net/http"
	"slices"
	"time"
)

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

func (srv *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	sets, err := srv.store.GetSettings()
	if err != nil {
		http.Error(w, "settings unavailable", 500)
		return
	}
	members, err := srv.store.GetMembers()
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

	meals, err := srv.store.GetMeals(dateStrs)
	if err != nil {
		http.Error(w, "meals unavailable", 500)
		return
	}

	rows := make([]MemberRow, len(members))
	for i, m := range members {
		cells := make([]MealCell, len(dates))
		memberTotal := 0
		for j, dc := range dates {
			chk := slices.Contains(meals[dc.Date], m.ID)
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

	srv.render(w, "index.html", IndexData{
		CurrentPage: "index",
		Members:     rows,
		Dates:       dates,
		Total:       grand,
		HasMembers:  len(rows) > 0,
	})
}
