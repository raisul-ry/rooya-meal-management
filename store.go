package main

// Member is a registered lunch participant.
type Member struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Settings holds configurable behaviour.
type Settings struct {
	DeadlineHour   int `json:"deadline_hour"`
	DeadlineMinute int `json:"deadline_minute"`
}

// Store abstracts all persistence. Swap JSON ↔ Postgres without touching handlers.
type Store interface {
	// Members
	GetMembers() ([]Member, error)
	AddMember(m Member) error
	MemberNameExists(name string) (bool, error)
	DeleteMember(id string) error

	// Meals — keys are "YYYY-MM-DD" strings, values are slices of member IDs.
	GetMeals(dates []string) (map[string][]string, error)
	GetPastMeals(before string) (map[string][]string, error)
	ToggleMeal(date, memberID string) (checked bool, dayTotal int, err error)

	// Settings
	GetSettings() (Settings, error)
	SaveSettings(s Settings) error
}
