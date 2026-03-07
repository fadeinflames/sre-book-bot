# SRE Learning Bot (Go, single-tenant)

Telegram-бот для поэтапного обучения SRE-основам с:
- учебными модулями и уроками,
- квизами и базовым mastery score,
- чеклистами и mentor review,
- SRS-повторением сложных тем,
- отчетами по прогрессу джунов.

## Контентные источники

Используются открытые и локальные источники:
- Google SRE Books: <https://sre.google/books/>
- The Site Reliability Workbook (англ., используется через русские конспекты внутри уроков)
- `SRE_Коллективный_разум.pdf` (локальный источник)
- `Site_Reliability_Engineering.pdf` (локальный источник)

Бот хранит структурированные конспекты и ссылки на источники, а не полные копии книг.

## Быстрый старт

1. Скопируйте `.env.example` в `.env` и заполните `BOT_TOKEN`.
2. (Опционально) добавьте allowlist:
   - `ALLOWED_TELEGRAM_IDS`
   - `MENTOR_TELEGRAM_IDS`
   - `ADMIN_TELEGRAM_IDS`
3. Запустите стек:

```bash
docker compose up --build
```

Сервисы:
- `api` на `http://localhost:8080`
- `bot` (long polling)
- `worker` (напоминания на повторение)
- `postgres`, `redis`
- `migrate` (однократное применение миграций)

## Основные команды бота

- `/start`
- `/roadmap`
- `/lesson`
- `/done <lesson_id>`
- `/quiz <module_slug>`
- `/review`
- `/progress`
- `/sources <module_slug>`
- `/checklists <module_slug>`
- `/submit_checklist <checklist_id> <notes>`
- `/mentor_report` (mentor/admin)
- `/pending_reviews` (mentor/admin)
- `/review_submission <id> approve|rework [comment]` (mentor/admin)

## API

- `GET /healthz`
- `GET /metrics`
- `GET /modules?telegram_id=<id>`
- `GET /users/progress?telegram_id=<id>`
- `GET /mentor/reports`
- `POST /mentor/checklists/review`

## Архитектура

- `cmd/bot` — Telegram bot runtime
- `cmd/api` — HTTP API/отчеты/health/metrics
- `cmd/worker` — SRS reminder worker
- `internal/app` — доменная логика + store
- `migrations` — схема и seed-контент

