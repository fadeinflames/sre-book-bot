-- Откат: возвращаем состояние после 0009 (дубликат B/C, correct C).
UPDATE quiz_questions
SET option_c = 'Секретность отчёта', correct_option = 'C'
WHERE question = 'Что обязательно в blameless postmortem?';
