-- Порционный текст урока: фрагменты для выдачи в боте
CREATE TABLE IF NOT EXISTS lesson_content_chunks (
    id BIGSERIAL PRIMARY KEY,
    lesson_id BIGINT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,
    body_text TEXT NOT NULL,
    UNIQUE(lesson_id, chunk_index)
);

CREATE INDEX IF NOT EXISTS idx_lesson_chunks_lesson ON lesson_content_chunks(lesson_id);

-- Прогресс чтения: какой фрагмент пользователь уже получил
CREATE TABLE IF NOT EXISTS user_lesson_reading (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id BIGINT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    last_chunk_index INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY(user_id, lesson_id)
);
