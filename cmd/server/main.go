package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/raisul-ry/rooya-meal-management/internal/store"
	"github.com/raisul-ry/rooya-meal-management/internal/web"
)

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

	srv := web.NewServer(s, os.Getenv("SETTINGS_PASSWORD"), baseDir)
	log.Println("Meal Manager → http://localhost:5050")
	log.Fatal(http.ListenAndServe(":5050", srv.Routes()))
}
