# ⚡ DevPulse (Developer Work Ledger & Standup Microservice)

<p align="center">
  <b>The Zero-Friction Developer Second Brain, Daily Standup & PKL Logbook Generator</b><br>
  <i>Automatic GitHub commit syncing + casual chat micro-journaling in one Telegram Bot.</i>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version" />
  <img src="https://img.shields.io/badge/PostgreSQL-15+-336791?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/Docker-Multi--Stage-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker" />
  <img src="https://img.shields.io/badge/Architecture-Clean%20%2F%20Zero--Slop-black?style=for-the-badge" alt="Architecture" />
</p>

---

## 🌟 Executive Summary & Concept

Banyak developer, anak magang, dan siswa PKL menghadapi masalah yang sama:
- **Lupa kemarin ngerjain apa** saat ditanya tim/lead di sesi *daily standup* pagi.
- **Pusing menyusun Logbook Mingguan PKL / Laporan Kerja** di akhir pekan karena catatan kerjaan tercecer di mana-mana.
- **Malas buka aplikasi manajemen proyek yang berat** (Jira/Notion) hanya untuk mencatat hal-hal kecil seperti diskusi meeting atau bugfix singkat.

**DevPulse** memecahkan masalah ini dengan konsep **Zero-Friction Work Logging**:
1. **Passive GitHub Ingestion**: Cukup daftarkan repositori yang Anda kerjakan (`/repo add <owner/repo>`). DevPulse otomatis menarik commit kodingan Anda dan mengarsipkannya secara cerdas.
2. **Casual Chat Micro-Journaling**: Cukup kirim chat santai di Telegram (contoh: `meeting sama pak andi bahas arsitektur database`, `fix bug cors reverse proxy satpam-nginx`). Parser otomatis mengklasifikasikan kategori (Meeting, Bugfix, Riset, Deploy, Koding).
3. **Automated Daily Standup (08:30 WIB)**: Setiap pagi hari kerja, bot otomatis mengirimkan draf format *Yesterday*, *Today*, dan *Blockers* yang siap di-copy ke grup tim.
4. **Automated Friday PKL Logbook (17:00 WIB)**: Setiap Jumat sore, bot merangkum seluruh aktivitas 5 hari kerja menjadi format tabel Logbook Mingguan siap cetak/lapor.

---

## 🚀 Fitur Utama

- **🐙 Multi-Repo Commit Tracking**: Pantau commit dari berbagai repo GitHub sekaligus dengan sistem pencegah duplikasi commit (*idempotent hash index*).
- **☕ Instant Standup Generator (`/standup`)**: Menghasilkan format standup profesional dalam bahasa Indonesia rapi kapan saja secara on-demand.
- **📋 Weekly PKL Logbook (`/logbook`)**: Format tabel hari Senin sampai Jumat lengkap dengan total capaian kategori koding, meeting, riset, dan deploy.
- **🎯 Task & Blocker Tracker (`/todo` & `/blocker`)**: Catat rencana hari ini dan kendala teknis (`blocker: nunggu token mentor`). Saat to-do diselesaikan via `/todo done <id>`, otomatis tercatat sebagai capaian kerja hari ini.
- **🔍 Natural Work Search (`/cari <keyword>`)**: Cari kembali catatan koding masa lalu dalam hitungan detik.
- **⚡ Super Lightweight**: Dibangun murni dengan Go standard library + driver atomic `pgx/v5` dan `robfig/cron/v3`. Tanpa framework berat eksternal.

---

## 💬 Contoh Interaksi Chat Telegram

### 1. Pencatatan Aktivitas Santai
> **User**: `meeting sama pak andi bahas arsitektur database`
```text
📝 Aktivitas Tercatat!
• Kategori: 👥 MEETING
• Catatan: Meeting sama pak andi bahas arsitektur database
• Tanggal: Rabu, 11 Nov 2026
```

> **User**: `fix bug cors di reverse proxy satpam-nginx`
```text
📝 Aktivitas Tercatat!
• Kategori: 🐛 BUGFIX
• Catatan: Fix bug cors di reverse proxy satpam-nginx
• Tanggal: Rabu, 11 Nov 2026
```

### 2. Draf Daily Standup Otomatis (`/standup` atau Pukul 08:30 WIB)
> **DevPulse Bot**:
```text
☕ Daily Standup — Rabu, 11 Nov 2026
👤 Engineer: @AryaKhaslana

⏮️ Kemarin (Yesterday):
• 💻 [AryaKhaslana/PocketPulse] feat(dashboard): add live Mini App dashboard
• 👥 Meeting sama pak andi bahas arsitektur database
• 🐛 Fix bug cors di reverse proxy satpam-nginx

🎯 Hari Ini (Today):
• Bikin unit test endpoint login
• Slicing UI dashboard

🚧 Kendala (Blockers):
• ⚠️ Nunggu access token cloudflare dari mentor
```

### 3. Logbook Mingguan PKL (`/logbook` atau Jumat 17:00 WIB)
> **DevPulse Bot**:
```text
📋 LOGBOOK MINGGUAN PKL / KERJA
📅 Periode: 09 Nov 2026 — 13 Nov 2026
👤 Praktikan/Engineer: @AryaKhaslana

1. Senin, 09 Nov 2026
   • 💻 [PocketPulse] feat(analytics): runway calculation
   • 👥 Meeting koordinasi mingguan tim

2. Selasa, 10 Nov 2026
   • 💻 [PocketPulse] feat(debt): talangan tracker & atomic settlement
   • 🔍 Riset query atomic postgres

3. Rabu, 11 Nov 2026
   • 💻 [DevPulse] Inisialisasi microservice git logger
   • 🚀 Deploy staging docker-compose di VPS

─────────────────────────────
📊 Total Capaian: 6 Aktivitas
• 💻 Koding/Git: 3
• 👥 Meeting/Koordinasi: 1
• 🐛 Bugfix: 0
• 🔍 Riset/Belajar: 1
• 🚀 Deploy/DevOps: 1
```

---

## 🎮 Command Cheatsheet

| Command | Parameter | Fungsi |
| :--- | :--- | :--- |
| `/start` | - | Inisialisasi akun & panduan cepat DevPulse |
| `/standup` | - | Buat draf Daily Standup hari ini secara instan |
| `/logbook` | - | Generate tabel Laporan Mingguan PKL (Senin - Jumat) |
| `/today` | - | Tampilkan timeline kronologis aktivitas hari ini |
| `/todo` | `[list / add <teks> / done <id>]` | Kelola daftar to-do & rencana kerja aktif |
| `/blocker` | `[list / add <teks> / done <id>]` | Kelola kendala / hambatan teknis aktif |
| `/github` | `set <username> [token]` | Set akun GitHub untuk tracking commit |
| `/repo` | `[list / add <owner/repo> / del <repo>]` | Kelola daftar repositori yang dipantau |
| `/sync` | - | Sinkronisasi commit terbaru GitHub sekarang |
| `/cari` | `<kata_kunci>` | Cari riwayat kodingan & memo masa lalu |
| `/catat` | `<kegiatan>` | Catat aktivitas kerja secara manual eksplisit |

---

## 🏗️ Arsitektur Direktori

```text
devpulse/
├── cmd/
│   └── bot/
│       └── main.go               # Entrypoint: DB pool, cron, GitHub client, bot dispatcher
├── internal/
│   ├── config/
│   │   └── config.go             # Environment loader (BOT_TOKEN, DATABASE_URL, TIMEZONE)
│   ├── database/
│   │   ├── db.go                 # pgx connection pool & embedded auto-migration
│   │   ├── db_test.go            # Integration test schema & table verification
│   │   └── migrations.sql        # Skema DDL ACID (Accounts, MonitoredRepos, WorkLogs, Todos, Blockers)
│   ├── domain/
│   │   ├── account.go            # Entity & Repository Interface Account
│   │   ├── blocker.go            # Entity & Repository Interface Blocker
│   │   ├── monitored_repo.go     # Entity & Repository Interface MonitoredRepo
│   │   ├── todo.go               # Entity & Repository Interface Todo
│   │   └── work_log.go           # Entity & Repository Interface WorkLog
│   ├── github/
│   │   ├── client.go             # GitHub API Client (Commits fetcher & repo validator)
│   │   └── client_test.go        # Unit test GitHub parser & mock client
│   ├── parser/
│   │   ├── parser.go             # NLP Classifier (Meeting, Bugfix, Riset, Deploy, Todo, Blocker)
│   │   └── parser_test.go        # Unit test variasi chat natural language
│   ├── synthesizer/
│   │   ├── logbook.go            # Generator Laporan Mingguan PKL format Senin-Jumat
│   │   ├── standup.go            # Generator Daily Standup (Yesterday, Today, Blockers)
│   │   └── synthesizer_test.go   # Unit test standup & logbook synthesis
│   ├── repository/
│   │   └── postgres/             # Query SQL relasional & batch commit deduplication
│   │       ├── account_repo.go
│   │       ├── blocker_repo.go
│   │       ├── monitored_repo_repo.go
│   │       ├── repo_test.go      # Integration test CRUD lengkap terhadap PostgreSQL
│   │       ├── todo_repo.go
│   │       └── work_log_repo.go
│   ├── scheduler/
│   │   ├── cron.go               # Scheduler Standup (08:30 WIB), Logbook (Jumat 17:00), & Git Sync
│   │   └── cron_test.go          # Unit test scheduler jobs
│   └── telegram/
│       ├── bot.go                # Client Telegram Bot API (Long polling / Webhook)
│       ├── handlers.go           # Dispatcher command vs casual chat logging
│       ├── views.go              # Renderer pesan & format Markdown
│       └── views_test.go         # Unit test renderer views
├── Dockerfile                    # Multi-stage static build (< 15MB final image)
├── docker-compose.yml            # Containerized DevPulse Bot + PostgreSQL 15 (port 5433)
├── .env.example
├── go.mod
└── go.sum
```

---

## 🛠️ Panduan Menjalankan

### 1. Salin Environment
```bash
cp .env.example .env
```
Isi `BOT_TOKEN` di file `.env` dengan token bot dari [@BotFather](https://t.me/botfather).

### 2. Jalankan via Docker Compose
```bash
docker compose up -d
```
Container database PostgreSQL dan microservice bot akan otomatis dibuat, dimigrasi, dan langsung mendengarkan pesan Telegram.

### 3. Cek Status & Log
```bash
# Cek container aktif
docker compose ps

# Cek log aplikasi
docker compose logs -f bot
```

### 4. Menjalankan Unit & Integration Test
```bash
go test -v ./...
```

---

## 📜 Lisensi
MIT License &copy; 2026 Arya Khaslana. Developed with clean architecture & zero-slop engineering standards.
