INSERT INTO modules (slug, title, description, order_index) VALUES
  ('foundations', 'SRE Foundations', 'База SRE: роль, цели, ответственность и связь с бизнесом.', 1),
  ('sli-slo-sla', 'SLI/SLO/SLA и Error Budget', 'Практика постановки надежности через измеримые цели.', 2),
  ('monitoring-alerting', 'Monitoring и Alerting', 'Наблюдаемость, золотые сигналы, burn-rate алерты.', 3),
  ('incident-response', 'Incident Response и Postmortem', 'Инциденты, on-call, postmortem и action items.', 4),
  ('toil-automation', 'Toil и Automation', 'Снижение рутины и автоматизация для масштабирования.', 5),
  ('capacity-release', 'Capacity и Release Reliability', 'Планирование нагрузки и безопасные релизы.', 6),
  ('kubernetes-reliability', 'Kubernetes Reliability', 'Практики надежности для k8s и платформенных систем.', 7)
ON CONFLICT (slug) DO NOTHING;

-- Lessons: concise RU summaries with source references from SRE Book + Workbook + local RU community book.
INSERT INTO lessons (module_id, title, content, order_index)
SELECT m.id, v.title, v.content, v.order_index
FROM modules m
JOIN (VALUES
  ('foundations', 'Что такое SRE и зачем', 'SRE = инженерный подход к эксплуатации. Фокус: надежность как продуктовая функция, измерение через SLO, баланс скорости и стабильности через error budget. По книге "SRE: Коллективный разум" глава 1 и SRE Book глава 1.', 1),
  ('foundations', 'SRE, DevOps, Platform Engineering', 'Сравнить роли и границы ответственности. Идея: DevOps как культура/коллаборация, SRE как практическая операционализация надежности с четкими метриками и ownership.', 2),

  ('sli-slo-sla', 'Как формулировать SLI', 'Определи пользовательские джорни, выбери 1-3 ключевых SLI (availability/latency/freshness), задавай только измеримые и бизнес-релевантные индикаторы.', 1),
  ('sli-slo-sla', 'SLO и Error Budget Policy', 'Задай SLO-цели, окно оценки и правила freeze/continue релизов. Используй policy на burn rate для действий команды.', 2),

  ('monitoring-alerting', 'Четыре золотых сигнала', 'Latency, Traffic, Errors, Saturation как базовый срез продакшн-состояния. Метрики должны вести к действию.', 1),
  ('monitoring-alerting', 'Alerting на burn rate', 'Используй multi-window/multi-burn-rate алерты, чтобы ловить и быстрые, и медленные деградации.', 2),

  ('incident-response', 'Инцидентный процесс', 'Роли в инциденте: incident commander, communications, ops lead. Таймлайн, эскалация, чаты, артефакты.', 1),
  ('incident-response', 'Blameless postmortem', 'Фиксируй контекст, impact, timeline, contributing factors, action items с владельцами и дедлайнами.', 2),

  ('toil-automation', 'Что считать Toil', 'Toil: ручная, повторяемая, реактивная работа без устойчивой ценности. Цель: ограничить долю toil и автоматизировать.', 1),
  ('toil-automation', 'Автоматизация как стратегия надежности', 'Автоматизируй runbook-процессы, проверки деплоев, rollback и стандартизированные операции.', 2),

  ('capacity-release', 'Capacity planning', 'Прогнозируй рост, определяй узкие места, тестируй отказоустойчивость и лимиты заранее.', 1),
  ('capacity-release', 'Надежные релизы', 'Progressive delivery, canary, rollback triggers, change budget и наблюдаемость во время релиза.', 2),

  ('kubernetes-reliability', 'Надежность в Kubernetes', 'Проверяй readiness/liveness, limits/requests, PDB, anti-affinity и отказоустойчивую топологию.', 1),
  ('kubernetes-reliability', 'Хаос и нагрузочное тестирование', 'Проводить controlled fault injection и нагрузочные тесты для валидации SLO-гипотез.', 2)
) AS v(slug, title, content, order_index) ON v.slug = m.slug
ON CONFLICT (module_id, order_index) DO NOTHING;

-- Resources include official SRE books page and specific book references
INSERT INTO resources (lesson_id, title, url)
SELECT l.id, r.title, r.url
FROM lessons l
JOIN (
  SELECT 'Что такое SRE и зачем'::text AS lesson_title, 'Google SRE Books'::text AS title, 'https://sre.google/books/'::text AS url
  UNION ALL SELECT 'Что такое SRE и зачем', 'SRE: Коллективный разум (книга RU)', 'https://sre.google/books/'
  UNION ALL SELECT 'SRE, DevOps, Platform Engineering', 'The Site Reliability Workbook - How SRE relates to DevOps', 'https://sre.google/workbook/how-sre-relates/'
  UNION ALL SELECT 'SRE, DevOps, Platform Engineering', 'The Site Reliability Workbook', 'https://sre.google/workbook/'
  UNION ALL SELECT 'Как формулировать SLI', 'Site Reliability Engineering Book - SLO Chapter', 'https://sre.google/sre-book/service-level-objectives/'
  UNION ALL SELECT 'SLO и Error Budget Policy', 'The Site Reliability Workbook - Implementing SLOs', 'https://sre.google/workbook/implementing-slos/'
  UNION ALL SELECT 'Четыре золотых сигнала', 'Site Reliability Engineering Book - Monitoring Distributed Systems', 'https://sre.google/sre-book/monitoring-distributed-systems/'
  UNION ALL SELECT 'Alerting на burn rate', 'The Site Reliability Workbook - Alerting on SLOs', 'https://sre.google/workbook/alerting-on-slos/'
  UNION ALL SELECT 'Инцидентный процесс', 'The Site Reliability Workbook - Incident Response', 'https://sre.google/workbook/incident-response/'
  UNION ALL SELECT 'Blameless postmortem', 'The Site Reliability Workbook - Postmortem Culture', 'https://sre.google/workbook/postmortem-culture/'
  UNION ALL SELECT 'Что считать Toil', 'Site Reliability Engineering Book - Eliminating Toil', 'https://sre.google/sre-book/eliminating-toil/'
  UNION ALL SELECT 'Автоматизация как стратегия надежности', 'Site Reliability Engineering Book - Automation', 'https://sre.google/sre-book/automation-at-google/'
  UNION ALL SELECT 'Надежные релизы', 'Site Reliability Engineering Book - Reliable Product Launches', 'https://sre.google/sre-book/reliable-product-launches/'
  UNION ALL SELECT 'Надежность в Kubernetes', 'Site Reliability Engineering Book - Managing critical state', 'https://sre.google/sre-book/'
  UNION ALL SELECT 'Хаос и нагрузочное тестирование', 'The Site Reliability Workbook - Testing reliability', 'https://sre.google/workbook/'
) r ON r.lesson_title = l.title
ON CONFLICT DO NOTHING;

INSERT INTO quiz_questions (module_id, question, option_a, option_b, option_c, option_d, correct_option, explanation, source_url)
SELECT m.id, q.question, q.a, q.b, q.c, q.d, q.correct, q.explanation, q.source
FROM modules m
JOIN (VALUES
  ('foundations', 'Какое ключевое отличие SRE-подхода?', 'Фокус только на инфраструктуре', 'Надёжность как инженерная и продуктовая функция', 'Полный отказ от релизов', 'Только ручные процессы', 'B', 'SRE связывает надёжность с инженерными практиками и бизнес-целями.', 'https://sre.google/books/'),
  ('foundations', 'Какой практический принцип SRE помогает балансировать скорость и стабильность?', 'Code freeze forever', 'Error budget', 'Нулевые инциденты любой ценой', 'Отмена on-call', 'B', 'Error budget делает риск управляемым и позволяет принимать решения о релизах.', 'https://sre.google/sre-book/embracing-risk/'),

  ('sli-slo-sla', 'Какое из перечисленного является SLI?', 'Обещание клиенту о штрафах', 'Время реакции поддержки', 'Измеряемая доля успешных запросов', 'Название SLA-документа', 'C', 'SLI — это измеримый индикатор качества сервиса.', 'https://sre.google/sre-book/service-level-objectives/'),
  ('sli-slo-sla', 'Когда полезно вводить error budget policy?', 'Только после аварии', 'Когда есть определённый SLO и его регулярный расчёт', 'Только при 99.999%', 'Никогда', 'B', 'Политика бюджета ошибок требует измеряемого SLO и режима принятия решений.', 'https://sre.google/workbook/implementing-slos/'),

  ('monitoring-alerting', 'Что входит в четыре золотых сигнала мониторинга?', 'CPU, RAM, Disk, Network', 'Latency, Traffic, Errors, Saturation', 'Only Logs and Traces', 'SLA, KPI, OKR, NPS', 'B', 'Классический набор сигналов для продакшн-мониторинга.', 'https://sre.google/sre-book/monitoring-distributed-systems/'),
  ('monitoring-alerting', 'Почему burn-rate алерты лучше статического порога?', 'Они всегда тише', 'Они учитывают скорость расходования error budget', 'Они не требуют SLO', 'Они только для low traffic', 'B', 'Burn-rate привязывает алерт к реальному риску нарушения SLO.', 'https://sre.google/workbook/alerting-on-slos/'),

  ('incident-response', 'Что обязательно в blameless postmortem?', 'Поиск виновного', 'Секретность отчёта', 'Root causes и action items с владельцами', 'Удаление логов', 'C', 'Цель постмортема — системные улучшения и исполнимые действия.', 'https://sre.google/workbook/postmortem-culture/'),
  ('incident-response', 'Что важнее в первые минуты инцидента?', 'Сразу писать отчёт', 'Назначить роли и стабилизировать сервис', 'Отключить мониторинг', 'Провести ретроспективу', 'B', 'Сначала управление инцидентом и снижение ущерба.', 'https://sre.google/workbook/incident-response/'),

  ('toil-automation', 'Какой признак у toil-работы?', 'Автоматизирована и масштабируема', 'Повторяемая ручная рутина без долговременной ценности', 'Ключевая инженерная разработка', 'Стратегическая архитектурная задача', 'B', 'Toil нужно снижать и автоматизировать.', 'https://sre.google/sre-book/eliminating-toil/'),
  ('toil-automation', 'Какой результат даёт правильная автоматизация?', 'Рост ручных операций', 'Снижение когнитивной нагрузки on-call и времени восстановления', 'Отмена SLO', 'Запрет релизов', 'B', 'Автоматизация повышает надёжность и скорость реакции.', 'https://sre.google/sre-book/automation-at-google/'),

  ('capacity-release', 'Что помогает снизить риск релиза?', 'Big bang deploy без метрик', 'Canary и rollback criteria', 'Отключение алертов', 'Релизы ночью без дежурных', 'B', 'Постепенный rollout и явные критерии отката — базовый паттерн надёжности.', 'https://sre.google/sre-book/reliable-product-launches/'),
  ('capacity-release', 'Зачем нужен capacity planning в SRE?', 'Только для отчётов', 'Чтобы заранее обнаруживать дефицит ресурсов и риски деградации', 'Чтобы уменьшить observability', 'Это не SRE-задача', 'B', 'Планирование нагрузки предотвращает аварийные дефициты.', 'https://sre.google/sre-book/capacity-planning/'),

  ('kubernetes-reliability', 'Какой набор повышает надёжность k8s-сервиса?', 'Отключить probes', 'Readiness/liveness + requests/limits + PDB', 'Только autoscaling', 'Только service mesh', 'B', 'Комбинация health checks и ограничений ресурсов критична для стабильности.', 'https://sre.google/books/'),
  ('kubernetes-reliability', 'Почему полезен chaos engineering?', 'Чтобы сломать прод без цели', 'Для проверки гипотез об отказоустойчивости до реальных аварий', 'Только для бенчмарков CPU', 'Он заменяет мониторинг', 'B', 'Контролируемые эксперименты выявляют слабые места заранее.', 'https://sre.google/workbook/')
) AS q(slug, question, a, b, c, d, correct, explanation, source) ON q.slug = m.slug
ON CONFLICT DO NOTHING;

INSERT INTO checklists (module_id, title, description)
SELECT m.id, c.title, c.description
FROM modules m
JOIN (VALUES
  ('sli-slo-sla', 'SLO Draft Checklist', 'Проверка готовности SLO перед запуском policy.'),
  ('monitoring-alerting', 'Alert Hygiene Checklist', 'Проверка качества алертов и сигналов.'),
  ('incident-response', 'Incident Readiness Checklist', 'Готовность команды к инцидентам и postmortem.'),
  ('toil-automation', 'Toil Reduction Checklist', 'План снижения рутины и автоматизации операций.')
) AS c(slug, title, description) ON c.slug = m.slug
ON CONFLICT DO NOTHING;

INSERT INTO checklist_items (checklist_id, item_text, order_index)
SELECT c.id, i.item_text, i.order_index
FROM checklists c
JOIN (VALUES
  ('SLO Draft Checklist', 'Определен пользовательский путь и его критичность.', 1),
  ('SLO Draft Checklist', 'Выбран SLI с понятным методом измерения.', 2),
  ('SLO Draft Checklist', 'Зафиксированы SLO-цели и окно оценки.', 3),
  ('SLO Draft Checklist', 'Согласована error budget policy с командой и стейкхолдерами.', 4),

  ('Alert Hygiene Checklist', 'Есть алерты на burn rate для fast/slow деградаций.', 1),
  ('Alert Hygiene Checklist', 'Каждый алерт имеет runbook и owner.', 2),
  ('Alert Hygiene Checklist', 'Регулярно удаляются noisy/non-actionable алерты.', 3),
  ('Alert Hygiene Checklist', 'Проверены каналы эскалации и on-call доступность.', 4),

  ('Incident Readiness Checklist', 'Назначены роли incident commander/comms/ops.', 1),
  ('Incident Readiness Checklist', 'Шаблон инцидента и postmortem стандартизирован.', 2),
  ('Incident Readiness Checklist', 'SLA коммуникаций с бизнесом согласован.', 3),
  ('Incident Readiness Checklist', 'Action items трекаются до закрытия.', 4),

  ('Toil Reduction Checklist', 'Идентифицированы top-5 toil задач за 30 дней.', 1),
  ('Toil Reduction Checklist', 'Для каждой toil-задачи есть решение automate/delete/delegate.', 2),
  ('Toil Reduction Checklist', 'Определены KPI снижения toil и целевые сроки.', 3),
  ('Toil Reduction Checklist', 'План автоматизации связан с on-call pain points.', 4)
) AS i(checklist_title, item_text, order_index) ON i.checklist_title = c.title
ON CONFLICT DO NOTHING;
