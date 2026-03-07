-- Исправление формулировок и грамматики в вопросах квиза (для уже применённого 0002)
UPDATE quiz_questions SET question = 'Какое ключевое отличие SRE-подхода?', option_b = 'Надёжность как инженерная и продуктовая функция', explanation = 'SRE связывает надёжность с инженерными практиками и бизнес-целями.' WHERE question = 'Что ключевое отличие SRE-подхода?';

UPDATE quiz_questions SET question = 'Какое из перечисленного является SLI?' WHERE question = 'Что из этого является SLI?';
UPDATE quiz_questions SET option_b = 'Когда есть определённый SLO и его регулярный расчёт' WHERE question = 'Когда полезно вводить error budget policy?';

UPDATE quiz_questions SET question = 'Что входит в четыре золотых сигнала мониторинга?' WHERE question = 'Что входит в "четыре золотых сигнала"?';

UPDATE quiz_questions SET option_c = 'Секретность отчёта', option_d = 'Root causes и action items с владельцами' WHERE question = 'Что обязательно в blameless postmortem?';
UPDATE quiz_questions SET option_a = 'Сразу писать отчёт' WHERE question = 'Что важнее в первые минуты инцидента?';

UPDATE quiz_questions SET question = 'Какой результат даёт правильная автоматизация?', explanation = 'Автоматизация повышает надёжность и скорость реакции.' WHERE question = 'Какой результат правильной автоматизации?';

UPDATE quiz_questions SET explanation = 'Постепенный rollout и явные критерии отката — базовый паттерн надёжности.' WHERE question = 'Что помогает снизить риск релиза?';
UPDATE quiz_questions SET question = 'Зачем нужен capacity planning в SRE?', option_a = 'Только для отчётов' WHERE question = 'Зачем capacity planning в SRE?';

UPDATE quiz_questions SET question = 'Какой набор повышает надёжность k8s-сервиса?', explanation = 'Комбинация health checks и ограничений ресурсов критична для стабильности.' WHERE question = 'Какой набор повышает надежность k8s сервиса?';
