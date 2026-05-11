<div align="center">

# 🍽️ Rooya Meal Manager

*Internal lunch sign-up tool for the Rooya team*

[![Live](https://img.shields.io/badge/Live-rooya--meal--management.onrender.com-059669?style=flat-square&logo=render&logoColor=white)](https://rooya-meal-management.onrender.com)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![Supabase](https://img.shields.io/badge/Database-Supabase-3ECF8E?style=flat-square&logo=supabase&logoColor=white)](https://supabase.com)
[![Render](https://img.shields.io/badge/Hosted%20on-Render-46E3B7?style=flat-square&logo=render&logoColor=white)](https://render.com)

</div>

---

## ✨ Features

- 📅 **Rolling 7-day dashboard** — shows yesterday, today, and the next 5 days
- 🔒 **Deadline enforcement** — checkboxes lock automatically after the configured cut-off time
- 👥 **Bulk member management** — add multiple members at once; password-protected
- 📋 **Meal history** — full log of past lunch days with per-day attendance
- 🍜 **Weekly menu** — upload and display a PDF menu
- 🛡️ **Admin password** — settings and member management require a password
- 📱 **Mobile-friendly** — card layout on small screens with a slide-in side nav
- 🌙 **Auto midnight refresh** — page reloads at midnight to shift the rolling window
- 🌏 **Timezone-aware** — configurable via `TZ` environment variable

---

## 🛠️ Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.26 |
| Web framework | Go standard library (`net/http`) |
| Templating | Go `html/template` |
| Database | Supabase (PostgreSQL) via `lib/pq` |
| Frontend | Vanilla HTML, CSS, JavaScript |
| Hosting | Render (free tier) |

---

## ⚙️ Environment Variables

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | Yes | Supabase PostgreSQL connection URI |
| `SETTINGS_PASSWORD` | Yes | Password to protect settings and member management |
| `TZ` | Recommended | Timezone for deadline calculation (e.g. `Asia/Dhaka`) |

---

## 🗄️ Database Schema

Schema lives in [`db/schema.sql`](db/schema.sql). Run it once in the Supabase SQL Editor:

- `members` — registered members
- `meal_signups` — per-member per-day sign-ups
- `app_settings` — deadline hour/minute configuration

---

## 📁 Project Structure

```
├── cmd/server/
│   └── main.go      — Entry point: server startup, store wiring
├── internal/
│   ├── store/
│   │   ├── store.go — Store interface, Member and Settings types
│   │   ├── json.go  — Local JSON fallback implementation
│   │   └── pg.go    — PostgreSQL (Supabase) implementation
│   └── web/
│       ├── server.go   — Server struct, router, shared helpers
│       ├── index.go    — Dashboard handler and view models
│       ├── history.go  — History handler and view models
│       ├── menu.go     — Menu handler and PDF upload
│       ├── settings.go — Settings handler and view models
│       ├── members.go  — Member add/delete API handlers
│       └── meals.go    — Meal toggle API handler
├── db/
│   └── schema.sql   — Database schema
├── templates/       — Server-rendered HTML templates
├── static/          — CSS, JavaScript, logo, menu PDF
├── render.yaml      — Render deployment configuration
└── go.mod
```

---

## 🚀 Local Development

**Prerequisites:** Go 1.26+

```bash
# Run with local JSON storage (no database needed)
go run ./cmd/server/

# Run with Supabase
DATABASE_URL="postgresql://..." SETTINGS_PASSWORD="..." TZ="Asia/Dhaka" go run ./cmd/server/
```

App runs at `http://localhost:5050`.

---

## ☁️ Deployment

Hosted on [Render](https://render.com) via `render.yaml`. Deploys automatically on every push to `main`. Set `DATABASE_URL`, `SETTINGS_PASSWORD`, and `TZ` in the Render dashboard under **Environment**.

---

<div align="center">

*Private — internal use at [Rooya](https://www.rooya.ai)*

</div>
