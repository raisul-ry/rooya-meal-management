# Rooya Meal Manager

Internal lunch sign-up tool for Rooya. Members register for daily lunches before a configurable daily deadline. Built with Go and backed by Supabase PostgreSQL.

**Live:** [rooya-meal-management.onrender.com](https://rooya-meal-management.onrender.com)

---

## Features

- **Rolling 7-day dashboard** — shows yesterday, today, and the next 5 days
- **Deadline enforcement** — checkboxes lock automatically after the configured cut-off time
- **Bulk member management** — add multiple members at once; password-protected
- **Meal history** — full log of past lunch days with per-day attendance
- **Weekly menu** — upload and display a PDF menu
- **Admin password** — settings and member management require a password
- **Mobile-friendly** — card layout on small screens with a slide-in side nav
- **Auto midnight refresh** — page reloads at midnight to shift the rolling window
- **Timezone-aware** — configurable via `TZ` environment variable

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.21 |
| Web framework | Go standard library (`net/http`) |
| Templating | Go `html/template` |
| Database | Supabase (PostgreSQL) via `lib/pq` |
| Frontend | Vanilla HTML, CSS, JavaScript |
| Hosting | Render (free tier) |

---

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | Yes (production) | Supabase PostgreSQL connection URI |
| `SETTINGS_PASSWORD` | Yes (production) | Password to protect settings and member management |
| `TZ` | Recommended | Timezone for deadline calculation (e.g. `Asia/Dhaka`) |

---

## Database Schema

Schema lives in [`db/schema.sql`](db/schema.sql). Run it once in the Supabase SQL Editor to create the required tables:

- `members` — registered members
- `meal_signups` — per-member per-day sign-ups
- `app_settings` — deadline hour/minute configuration

---

## Project Structure

```
├── main.go          — HTTP server, route handlers, view models
├── store.go         — Store interface and shared types
├── store_json.go    — Local JSON fallback implementation
├── store_pg.go      — PostgreSQL (Supabase) implementation
├── db/
│   └── schema.sql   — Database schema
├── templates/       — Server-rendered HTML templates
├── static/          — CSS, JavaScript, logo, menu PDF
├── render.yaml      — Render deployment configuration
└── go.mod
```

---

## Local Development

**Prerequisites:** Go 1.21+

```bash
# Run with local JSON storage (no database needed)
go run .

# Run with Supabase
DATABASE_URL="postgresql://..." SETTINGS_PASSWORD="..." TZ="Asia/Dhaka" go run .
```

App runs at `http://localhost:5050`.

---

## Deployment

Hosted on [Render](https://render.com) via `render.yaml`. Deploys automatically on every push to `main`. Set `DATABASE_URL`, `SETTINGS_PASSWORD`, and `TZ` in the Render dashboard under **Environment**.

---

## License

Private — internal use only.
