package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/raisul-ry/rooya-meal-management/internal/notify"
	"github.com/raisul-ry/rooya-meal-management/internal/store"
	"github.com/raisul-ry/rooya-meal-management/internal/web"
)

func runScheduler(s store.Store) {
	for {
		settings, err := s.GetSettings()
		if err != nil {
			time.Sleep(time.Minute)
			continue
		}

		now := time.Now()
		base := time.Date(now.Year(), now.Month(), now.Day(),
			settings.DeadlineHour, settings.DeadlineMinute, 0, 0, time.Local)
		trigger := base.Add(15 * time.Minute)
		if !trigger.After(now) {
			trigger = trigger.AddDate(0, 0, 1)
		}

		time.Sleep(time.Until(trigger))

		settings, err = s.GetSettings()
		if err != nil || settings.TeamsWebhookURL == "" {
			continue
		}

		tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
		meals, err := s.GetMeals([]string{tomorrow})
		if err != nil {
			log.Printf("scheduler: get meals: %v", err)
			continue
		}
		count := len(meals[tomorrow])
		if err := notify.SendMealCount(settings.TeamsWebhookURL, tomorrow, count); err != nil {
			log.Printf("scheduler: teams notify: %v", err)
		} else {
			log.Printf("scheduler: sent meal count %d for %s", count, tomorrow)
		}
	}
}

func main() {
	if tz := os.Getenv("TZ"); tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			time.Local = loc
		}
	}

	baseDir, _ := os.Getwd()

	var s store.Store
	if connStr := os.Getenv("DATABASE_URL"); connStr != "" {
		pg, err := store.NewPGStore(connStr)
		if err != nil {
			log.Fatalf("postgres: %v", err)
		}
		log.Println("storage: PostgreSQL")
		s = pg
	} else {
		log.Println("storage: local JSON files")
		s = store.NewJSONStore(
			filepath.Join(baseDir, "data", "meals.json"),
			filepath.Join(baseDir, "data", "settings.json"),
		)
	}

	go runScheduler(s)

	srv := web.NewServer(s, os.Getenv("SETTINGS_PASSWORD"), baseDir)
	log.Println("Meal Manager → http://localhost:5050")
	log.Fatal(http.ListenAndServe(":5050", srv.Routes()))
}
