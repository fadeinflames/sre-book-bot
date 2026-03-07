-- Выжимки по всем урокам: бот как основной источник, без обязательного перехода во внешние материалы.
-- Урок "Что такое SRE и зачем" уже заполнен в 0004.

-- foundations: SRE, DevOps, Platform Engineering (dollar-quoting для многострочного текста)
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, 1, $$Ключевые отличия ролей:

DevOps — культура и практики: слом стены между разработкой и эксплуатацией, общие цели, общий инструментарий. Фокус на скорости доставки изменений.

SRE — конкретная реализация: инженеры с кодом, метрики надежности (SLO, error budget), явный ownership за прод. Фокус на надёжности и балансе с фиче-разработкой. «SRE — это то, что получится, если попросить software engineer спроектировать операционную функцию» (Google).

Platform Engineering — внутренние платформы и самообслуживание для разработчиков (пайплайны, среды, API). Может включать SRE-практики для платформы.

В реальности названия в компаниях смешиваются. Важно зафиксировать в команде: кто за что отвечает, как измеряем надёжность и как принимаем решения о релизах.$$
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'foundations'
WHERE l.title = 'SRE, DevOps, Platform Engineering'
ON CONFLICT (lesson_id, chunk_index) DO NOTHING;

-- sli-slo-sla: Как формулировать SLI
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, v.chunk_index, v.body_text
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'sli-slo-sla'
JOIN (VALUES
  (1, 'SLI (Service Level Indicator) — измеримый индикатор качества сервиса с точки зрения пользователя. Не «uptime сервера», а «доля успешных запросов» или «доля запросов быстрее 200 ms».

Как формулировать: 1) Определи пользовательский путь (что делает пользователь). 2) Выбери 1–3 ключевых SLI: доступность (успешность запросов), задержка (латентность перцентили), свежесть данных если применимо. 3) Задавай только то, что можно измерить инструментально и что реально влияет на опыт.'),
  (2, 'Примеры SLI: «доля HTTP-запросов с кодом 5xx от всех запросов», «доля запросов с latency < 200 ms (p99)», «доля задач в очереди, обработанных за 24 ч». Избегай «средней» доступности без перцентилей — хвост латентности часто важнее среднего.')
) AS v(chunk_index, body_text) ON true
WHERE l.title = 'Как формулировать SLI'
ON CONFLICT (lesson_id, chunk_index) DO NOTHING;

-- sli-slo-sla: SLO и Error Budget Policy
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, v.chunk_index, v.body_text
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'sli-slo-sla'
JOIN (VALUES
  (1, 'SLO (Service Level Objective) — целевое значение SLI, например «99.9% запросов успешны за месяц». SLA — договор с клиентом (часто с штрафами); SLO обычно жёстче SLA, чтобы был запас. Error budget = 1 − SLO (например 0.1% времени можно «тратить» на сбои).'),
  (2, 'Error budget policy: правила, что делать при расходовании бюджета. Примеры: исчерпан бюджет за месяц → freeze релизов до следующего периода; быстрый burn rate → алерт и разбор. Так команда балансирует скорость и стабильность без бесконечных споров. Зафиксируй окно (28–30 дней), пороги и кто принимает решение (релиз-менеджер, команда).')
) AS v(chunk_index, body_text) ON true
WHERE l.title = 'SLO и Error Budget Policy'
ON CONFLICT (lesson_id, chunk_index) DO NOTHING;

-- monitoring-alerting: Четыре золотых сигнала
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, v.chunk_index, v.body_text
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'monitoring-alerting'
JOIN (VALUES
  (1, 'Четыре золотых сигнала мониторинга (Google SRE): 1) Latency — задержка запросов (успешных и неуспешных отдельно). 2) Traffic — нагрузка (RPS, запросы в очередь). 3) Errors — доля ошибок (5xx, таймауты, отказы зависимостей). 4) Saturation — насколько ресурс «полон» (CPU, память, диск, квоты).'),
  (2, 'Правило: каждый сигнал должен вести к действию. Если метрика не влияет на решение «что делать» — не алертить по ней. Мониторинг методом «белого ящика» (внутри приложения) даёт причины; «чёрного ящика» (снаружи) — симптомы. Нужны оба. Следи за хвостом распределения (p99, p99.9), не только за средним.')
) AS v(chunk_index, body_text) ON true
WHERE l.title = 'Четыре золотых сигнала'
ON CONFLICT (lesson_id, chunk_index) DO NOTHING;

-- monitoring-alerting: Alerting на burn rate (dollar-quoting для многострочного текста)
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, 1, $$Алертить по burn rate выгоднее, чем по статическому порогу: «ошибок больше 1%» не учитывает, как быстро мы тратим error budget. Burn rate — во сколько раз быстрее расходуется бюджет (например, за 1 час израсходовали то, что по SLO можно было тратить 30 дней).

Подходы: 1) Один алерт «burn rate ≥ 1» за скользящее окно. 2) Multi-window, multi-burn-rate: быстрый burn за короткое окно и медленный за длинное — ловит и внезапные всплески, и ползучую деградацию. 3) Низкий трафик: искусственный трафик, объединение сервисов или смягчение SLO/окна.$$
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'monitoring-alerting'
WHERE l.title = 'Alerting на burn rate'
ON CONFLICT (lesson_id, chunk_index) DO NOTHING;

-- incident-response: Инцидентный процесс
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, v.chunk_index, v.body_text
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'incident-response'
JOIN (VALUES
  (1, 'Роли в инциденте: Incident Commander (один человек, координирует, не обязательно чинит сам). Communications — обновления для стейкхолдеров и пользователей. Ops/Technical lead — те, кто роют логи и перезапускают сервисы. Важно: роли зафиксированы, эскалация по таймеру, один канал (чат, колл).'),
  (2, 'Артефакты: таймлайн (когда что произошло), список действий и гипотез, лог решений. После стабилизации — постмортем без поиска виновных, с root causes и action items. В первые минуты приоритет — стабилизация и коммуникация, не расследование.')
) AS v(chunk_index, body_text) ON true
WHERE l.title = 'Инцидентный процесс'
ON CONFLICT (lesson_id, chunk_index) DO NOTHING;

-- incident-response: Blameless postmortem
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, 1, 'Blameless postmortem — разбор инцидента без назначения вины. Цель: понять систему и процессы, а не наказать человека. Содержимое: контекст и impact, таймлайн, факторы (что способствовало), корневые причины, action items с владельцами и сроками. Action items обязательны — иначе постмортем бесполезен. Не заводи KPI по «закрытию action items» без контекста: команды могут занижать сложность задач.'
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'incident-response'
WHERE l.title = 'Blameless postmortem'
ON CONFLICT (lesson_id, chunk_index) DO NOTHING;

-- toil-automation: Что считать Toil
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, 1, 'Toil — ручная, повторяемая, реактивная работа без долговременной ценности. Примеры: ручной перезапуск сервисов, разовые правки в конфигах, ответы на однотипные тикеты. Цель SRE — ограничить долю toil (например не более 50% времени) и сокращать его автоматизацией или устранением. Если toil растёт, команда не успевает заниматься инженерными задачами и улучшением системы.'
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'toil-automation'
WHERE l.title = 'Что считать Toil'
ON CONFLICT (lesson_id, chunk_index) DO NOTHING;

-- toil-automation: Автоматизация как стратегия надежности
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, 1, 'Автоматизация снижает время восстановления и когнитивную нагрузку на дежурных. Что автоматизировать в приоритете: runbook-шаги (перезапуск, откат), проверки перед/после деплоя, откат по критериям (например по SLO). Идемпотентность и тесты на окружении близком к проду уменьшают ошибки. «Automate this year''s job away» — постоянное смещение рутины в код.'
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'toil-automation'
WHERE l.title = 'Автоматизация как стратегия надежности'
ON CONFLICT (lesson_id, chunk_index) DO NOTHING;

-- capacity-release: Capacity planning
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, 1, 'Capacity planning — прогнозирование нагрузки и ресурсов, чтобы не уйти в дефицит неожиданно. Шаги: сбор метрик использования и роста, определение узких мест (CPU, память, сеть, лимиты внешних API), тесты на отказоустойчивость и лимиты. Планируй с запасом и пересматривай регулярно. Игнорирование ведёт к авариям при росте трафика или сбоях одного из компонентов.'
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'capacity-release'
WHERE l.title = 'Capacity planning'
ON CONFLICT (lesson_id, chunk_index) DO NOTHING;

-- capacity-release: Надежные релизы
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, 1, 'Надёжный релиз: постепенный rollout (canary, процент трафика), явные критерии отката (SLO, error rate, latency), наблюдаемость во время релиза. Change budget можно привязать к error budget. Автоматический откат при срабатывании алертов снижает время простоя. Релизы должны быть частыми и маленькими — так проще изолировать причину проблем.'
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'capacity-release'
WHERE l.title = 'Надежные релизы'
ON CONFLICT (lesson_id, chunk_index) DO NOTHING;

-- kubernetes-reliability: Надежность в Kubernetes
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, 1, 'Базовый набор для надёжности в k8s: Readiness probe (когда подать отдавать трафик) и Liveness probe (когда перезапустить контейнер). Requests и Limits для CPU/памяти — чтобы планировщик и kubelet не перегружали ноды. PodDisruptionBudget (PDB) — сколько подов могут быть недоступны при добровольном выводе нод. Anti-affinity и топология — распределение реплик по зонам/нодам для отказоустойчивости.'
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'kubernetes-reliability'
WHERE l.title = 'Надежность в Kubernetes'
ON CONFLICT (lesson_id, chunk_index) DO NOTHING;

-- kubernetes-reliability: Хаос и нагрузочное тестирование
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, 1, 'Хаос-инжиниринг — контролируемое внесение сбоев (убийство пода, отключение сети, задержки) для проверки устойчивости системы до того, как это произойдёт в проде. Нагрузочное тестирование проверяет поведение под пиковой нагрузкой и лимиты. Оба подхода валидируют гипотезы об SLO и отказоустойчивости. Начинай с тестовых окружений и чётких критериев остановки эксперимента.'
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'kubernetes-reliability'
WHERE l.title = 'Хаос и нагрузочное тестирование'
ON CONFLICT (lesson_id, chunk_index) DO NOTHING;
