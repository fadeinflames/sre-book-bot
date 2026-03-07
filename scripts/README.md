# Скрипты и утилиты для контента бота

## Go: pdf2chunks (без Python)

Утилита на Go извлекает текст через **pdftotext** (poppler) и выводит SQL для `lesson_content_chunks`. Зависимостей проекта хватает, Python не нужен.

**Установка pdftotext (один раз):**
```bash
brew install poppler   # macOS
```

**Запуск:**
```bash
# PDF передаётся как аргумент — утилита сама вызовет pdftotext
go run ./cmd/pdf2chunks --lesson-id 1 --pdf path/to/SRE_Коллективный_разум.pdf

# Или пайпом (если pdftotext уже запущен)
pdftotext -layout - path/to/book.pdf | go run ./cmd/pdf2chunks --lesson-id 2
```

Опции: `--chunk-size 3500` (по умолчанию), `--lesson-id N` (обязателен для SQL).

**Уже выгружено:** миграция `0006_seed_book_chunks.up.sql` заполняет уроки 1–4 текстом из книги «SRE: Коллективный разум» (главы 1–3). После `migrate up` бот отдаёт эти порции в `/lesson` и `/lesson_next` без ручного запуска скриптов.

---

## Python: extract_pdf_chunks.py (опционально)

Альтернатива: извлечение через pypdf, вывод SQL или JSON.

**Установка:** `pip install pypdf` или `pip install -r scripts/requirements.txt`

**Запуск:**
```bash
python scripts/extract_pdf_chunks.py path/to/book.pdf --lesson-id 1 --output sql
python scripts/extract_pdf_chunks.py path/to/book.pdf --output json
```

Чанки режутся по абзацам, лимит ~3500 символов.
