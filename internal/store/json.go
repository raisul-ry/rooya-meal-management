package store

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"sync"
)

type jsonData struct {
	Members []Member            `json:"members"`
	Meals   map[string][]string `json:"meals"`
}

// JSONStore persists data in two local JSON files.
type JSONStore struct {
	mu           sync.RWMutex
	mealsFile    string
	settingsFile string
}

func NewJSONStore(mealsFile, settingsFile string) *JSONStore {
	return &JSONStore{mealsFile: mealsFile, settingsFile: settingsFile}
}

func (s *JSONStore) read() jsonData {
	d := jsonData{Members: []Member{}, Meals: map[string][]string{}}
	b, err := os.ReadFile(s.mealsFile)
	if err != nil {
		return d
	}
	if err := json.Unmarshal(b, &d); err != nil {
		return d
	}
	if d.Members == nil {
		d.Members = []Member{}
	}
	if d.Meals == nil {
		d.Meals = map[string][]string{}
	}
	return d
}

func (s *JSONStore) write(d jsonData) error {
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.mealsFile, b, 0o644)
}

func (s *JSONStore) GetMembers() ([]Member, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.read().Members, nil
}

func (s *JSONStore) AddMember(m Member) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	d := s.read()
	d.Members = append(d.Members, m)
	return s.write(d)
}

func (s *JSONStore) MemberNameExists(name string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, m := range s.read().Members {
		if strings.EqualFold(m.Name, name) {
			return true, nil
		}
	}
	return false, nil
}

func (s *JSONStore) DeleteMember(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	d := s.read()
	d.Members = slices.DeleteFunc(d.Members, func(m Member) bool { return m.ID == id })
	for date := range d.Meals {
		d.Meals[date] = slices.DeleteFunc(d.Meals[date], func(mid string) bool { return mid == id })
	}
	return s.write(d)
}

func (s *JSONStore) GetMeals(dates []string) (map[string][]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d := s.read()
	result := make(map[string][]string, len(dates))
	for _, date := range dates {
		result[date] = d.Meals[date]
	}
	return result, nil
}

func (s *JSONStore) GetPastMeals(before string) (map[string][]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := map[string][]string{}
	for date, mids := range s.read().Meals {
		if date < before && len(mids) > 0 {
			result[date] = mids
		}
	}
	return result, nil
}

func (s *JSONStore) ToggleMeal(date, memberID string) (bool, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d := s.read()
	list := d.Meals[date]
	checked := false
	if slices.Contains(list, memberID) {
		d.Meals[date] = slices.DeleteFunc(list, func(mid string) bool { return mid == memberID })
	} else {
		d.Meals[date] = append(list, memberID)
		checked = true
	}
	return checked, len(d.Meals[date]), s.write(d)
}

func (s *JSONStore) GetSettings() (Settings, error) {
	defaults := Settings{DeadlineHour: 22, DeadlineMinute: 0}
	b, err := os.ReadFile(s.settingsFile)
	if err != nil {
		return defaults, nil
	}
	var set Settings
	if err := json.Unmarshal(b, &set); err != nil {
		return defaults, nil
	}
	return set, nil
}

func (s *JSONStore) SaveSettings(set Settings) error {
	b, _ := json.MarshalIndent(set, "", "  ")
	return os.WriteFile(s.settingsFile, b, 0o644)
}
