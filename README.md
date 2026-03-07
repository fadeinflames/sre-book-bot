# SRE Learning Bot (Go, single-tenant)

Telegram-бот для поэтапного обучения SRE-основам с:
- учебными модулями и уроками,
- квизами и базовым mastery score,
- чеклистами и mentor review,
- SRS-повторением сложных тем,
- отчетами по прогрессу джунов.

## Контент: бот как основной источник

Уроки выдаются порционно через `/lesson` и `/lesson_next`. Текст хранится в БД (таблица `lesson_content_chunks`) — выжимки по всем темам, чтобы джунам не обязательно идти во внешние источники. Ссылки в `/sources` остаются дополнительными.

Добавление текста из книг:
- Вручную: правка миграций или `INSERT` в `lesson_content_chunks`.
- Из PDF: скрипт `scripts/extract_pdf_chunks.py` вытаскивает текст из PDF и режет на чанки; можно вывести SQL под нужный `lesson_id`. См. `scripts/README.md`.

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
- `/lesson` — следующий урок (порциями)
- `/lesson_next` — следующая порция текущего урока
- `/done (lesson_id)`
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
- `migrations` — схема и seed-контент (в т.ч. выжимки по урокам)
- `scripts/` — выгрузка текста из PDF в чанки (см. `scripts/README.md`)

