package store

import (
	"database/sql"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

// PGStore persists data in a PostgreSQL database (Supabase-compatible).
type PGStore struct {
	db *sql.DB
}

func NewPGStore(connStr string) (*PGStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PGStore{db: db}, nil
}

func (s *PGStore) GetMembers() ([]Member, error) {
	rows, err := s.db.Query(`SELECT id::text, name FROM members ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Name); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	if members == nil {
		members = []Member{}
	}
	return members, rows.Err()
}

func (s *PGStore) AddMember(m Member) error {
	_, err := s.db.Exec(
		`INSERT INTO members (id, name) VALUES ($1::uuid, $2)`,
		m.ID, m.Name,
	)
	return err
}

func (s *PGStore) MemberNameExists(name string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM members WHERE lower(name) = lower($1))`,
		name,
	).Scan(&exists)
	return exists, err
}

func (s *PGStore) DeleteMember(id string) error {
	// meal_signups rows cascade-delete via FK constraint
	_, err := s.db.Exec(`DELETE FROM members WHERE id = $1::uuid`, id)
	return err
}

func (s *PGStore) GetMeals(dates []string) (map[string][]string, error) {
	result := make(map[string][]string, len(dates))
	for _, d := range dates {
		result[d] = []string{}
	}
	if len(dates) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(dates))
	args := make([]any, len(dates))
	for i, d := range dates {
		placeholders[i] = "$" + strconv.Itoa(i+1) + "::date"
		args[i] = d
	}
	query := `SELECT meal_date::text, member_id::text FROM meal_signups WHERE meal_date IN (` +
		strings.Join(placeholders, ",") + `)`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var date, mid string
		if err := rows.Scan(&date, &mid); err != nil {
			return nil, err
		}
		result[date] = append(result[date], mid)
	}
	return result, rows.Err()
}

func (s *PGStore) GetPastMeals(before string) (map[string][]string, error) {
	rows, err := s.db.Query(
		`SELECT meal_date::text, member_id::text FROM meal_signups WHERE meal_date < $1::date ORDER BY meal_date DESC`,
		before,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string][]string{}
	for rows.Next() {
		var date, mid string
		if err := rows.Scan(&date, &mid); err != nil {
			return nil, err
		}
		result[date] = append(result[date], mid)
	}
	return result, rows.Err()
}

func (s *PGStore) ToggleMeal(date, memberID string) (bool, int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`DELETE FROM meal_signups WHERE member_id = $1::uuid AND meal_date = $2::date`,
		memberID, date,
	)
	if err != nil {
		return false, 0, err
	}
	affected, _ := res.RowsAffected()
	checked := false
	if affected == 0 {
		_, err = tx.Exec(
			`INSERT INTO meal_signups (member_id, meal_date) VALUES ($1::uuid, $2::date)`,
			memberID, date,
		)
		if err != nil {
			return false, 0, err
		}
		checked = true
	}

	var total int
	if err = tx.QueryRow(
		`SELECT count(*) FROM meal_signups WHERE meal_date = $1::date`, date,
	).Scan(&total); err != nil {
		return false, 0, err
	}
	return checked, total, tx.Commit()
}

func (s *PGStore) GetSettings() (Settings, error) {
	var set Settings
	err := s.db.QueryRow(
		`SELECT deadline_hour, deadline_minute FROM app_settings WHERE id = 1`,
	).Scan(&set.DeadlineHour, &set.DeadlineMinute)
	if err == sql.ErrNoRows {
		return Settings{DeadlineHour: 22, DeadlineMinute: 0}, nil
	}
	return set, err
}

func (s *PGStore) SaveSettings(set Settings) error {
	_, err := s.db.Exec(`
		INSERT INTO app_settings (id, deadline_hour, deadline_minute)
		VALUES (1, $1, $2)
		ON CONFLICT (id) DO UPDATE SET deadline_hour = $1, deadline_minute = $2
	`, set.DeadlineHour, set.DeadlineMinute)
	return err
}
