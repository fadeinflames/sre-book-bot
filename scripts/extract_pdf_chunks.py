#!/usr/bin/env python3
"""
Выгрузка текста из PDF порциями для lesson_content_chunks.
Режет по абзацам, макс. ~3500 символов на чанк (лимит сообщения в Telegram).
"""
import argparse
import json
import re
import sys
from pathlib import Path

try:
    from pypdf import PdfReader
except ImportError:
    print("Установите pypdf: pip install pypdf", file=sys.stderr)
    sys.exit(1)

MAX_CHUNK = 3500


def extract_text(pdf_path: Path) -> str:
    reader = PdfReader(str(pdf_path))
    parts = []
    for page in reader.pages:
        parts.append(page.extract_text() or "")
    return "\n".join(parts)


def clean(text: str) -> str:
    text = re.sub(r"\n{3,}", "\n\n", text)
    text = re.sub(r" +\n", "\n", text)
    return text.strip()


def split_into_chunks(text: str, max_len: int = MAX_CHUNK) -> list[str]:
    text = clean(text)
    chunks = []
    current = []
    current_len = 0
    for para in text.split("\n\n"):
        para = para.strip()
        if not para:
            continue
        need = len(para) + 2
        if current_len + need > max_len and current:
            chunks.append("\n\n".join(current))
            current = []
            current_len = 0
        current.append(para)
        current_len += need
    if current:
        chunks.append("\n\n".join(current))
    return chunks


def escape_sql(s: str) -> str:
    return s.replace("'", "''").replace("\\", "\\\\")


def main() -> None:
    ap = argparse.ArgumentParser(description="Extract PDF text into chunks for lesson_content_chunks")
    ap.add_argument("pdf", type=Path, help="Path to PDF file")
    ap.add_argument("--lesson-id", type=int, default=None, help="Lesson ID for SQL output")
    ap.add_argument("--output", choices=("sql", "json"), default="json", help="Output format")
    ap.add_argument("--max-chunk", type=int, default=MAX_CHUNK, help="Max characters per chunk")
    args = ap.parse_args()

    if not args.pdf.exists():
        print(f"File not found: {args.pdf}", file=sys.stderr)
        sys.exit(1)

    raw = extract_text(args.pdf)
    if not raw.strip():
        print("No text extracted from PDF.", file=sys.stderr)
        sys.exit(1)

    chunks = split_into_chunks(raw, max_len=args.max_chunk)

    if args.output == "json":
        print(json.dumps({"chunks": chunks, "total": len(chunks)}, ensure_ascii=False, indent=2))
        return

    if args.lesson_id is None:
        print("For SQL output specify --lesson-id", file=sys.stderr)
        sys.exit(1)

    print("-- Paste into a migration or run in psql")
    print("INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text) VALUES")
    lines = []
    for i, c in enumerate(chunks, 1):
        body = escape_sql(c)
        lines.append(f"  ({args.lesson_id}, {i}, '{body}')")
    print(",\n".join(lines))
    print("ON CONFLICT (lesson_id, chunk_index) DO UPDATE SET body_text = EXCLUDED.body_text;")


if __name__ == "__main__":
    main()
