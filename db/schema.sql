-- ============================================================
-- Rooya Meal Manager — Supabase / PostgreSQL schema
-- Run this in the Supabase SQL Editor (Dashboard → SQL Editor)
-- ============================================================

-- Members
CREATE TABLE IF NOT EXISTS members (
    id          UUID        PRIMARY KEY,
    name        TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Case-insensitive unique constraint on name
CREATE UNIQUE INDEX IF NOT EXISTS members_name_lower_idx
    ON members (lower(name));

-- Meal sign-ups  (composite PK prevents double-booking)
CREATE TABLE IF NOT EXISTS meal_signups (
    member_id   UUID        NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    meal_date   DATE        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (member_id, meal_date)
);

CREATE INDEX IF NOT EXISTS meal_signups_date_idx ON meal_signups (meal_date);

-- Application settings  (enforced single row via CHECK id = 1)
CREATE TABLE IF NOT EXISTS app_settings (
    id              INTEGER PRIMARY KEY DEFAULT 1,
    deadline_hour   INTEGER NOT NULL DEFAULT 22,
    deadline_minute INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT single_row CHECK (id = 1)
);

-- Seed default settings row
INSERT INTO app_settings (id, deadline_hour, deadline_minute)
VALUES (1, 22, 0)
ON CONFLICT (id) DO NOTHING;
