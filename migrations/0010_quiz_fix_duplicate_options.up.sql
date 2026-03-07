-- Исправление дублирования вариантов ответа в вопросе про blameless postmortem.
-- После 0009 option_b и option_c оба стали "Секретность отчёта"; правильный ответ
-- в option_d, но correct_option остался 'C'. Восстанавливаем четыре разных варианта.
UPDATE quiz_questions
SET option_c = 'Удаление логов', correct_option = 'D'
WHERE question = 'Что обязательно в blameless postmortem?';
