-- Длинный текст из книги «SRE: Коллективный разум» для уроков 5–14 (мониторинг, алертинг, инциденты, постмортемы, toil, capacity, релизы, k8s, хаос).

-- Урок 5: Четыре золотых сигнала (Глава 5 Мониторинг)
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, v.chunk_index, v.body_text
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'monitoring-alerting'
JOIN (VALUES
  (1, 'Мониторинг — фундамент SRE. Без мониторинга невозможно определить надёжность, рассчитать error budget, принимать решения о выкатке или заморозке. Между «у нас есть мониторинг» и «мы понимаем, что происходит» — пропасть. USE (Utilization, Saturation, Errors) — для исчерпаемых ресурсов: CPU, память, диски, пулы коннекшенов. RED (Request rate, Error rate, Duration) — для сервисов: RPS, доля ошибок, латенси. USE и RED дополняют друг друга: пул коннекшенов к БД — ресурс, ему нужны USE-метрики; вызов API — запрос, ему нужны RED. Без USE не увидите исчерпание пула до таймаутов. Метрики зависимостей: для каждой исходящей зависимости (БД, API, кэш) — RED. При алерте открываете дашборд зависимостей и за 10–20 секунд видите деградировавшую зависимость.'),
  (2, 'Диагностика через очереди: рост очереди — опережающий индикатор. Очередь растёт до роста latency и error rate. Идентифицируйте все очереди (пулы потоков, connection pools, Kafka), мониторьте длину и скорость поступления/обработки. Алерт при аномальном росте. При инциденте первое действие — посмотреть, где растут очереди. Аномалии: при сезонности (ночь/день) статический порог не подходит; медиана за «сейчас ±час» и 1/2/3 недели назад в тот же слот даёт устойчивую оценку. Где мерить: whitebox с сервиса — точные данные; blackbox (probes) — пользовательский опыт, маршрутизация, DNS. Решение: двухуровневый мониторинг. Дашборд: аннотации деплоев, feature flags, инциденты на одном графике с SLO. Три уровня: обзорный (все сервисы), сервисный (RED + USE + зависимости), debug. «Мониторинг на базе генерального директора»: сначала узнаём о проблеме по звонку гендира, потом одновременно, потом раньше всех. Цель — узнавать раньше всех.')
) AS v(chunk_index, body_text) ON true
WHERE l.title = 'Четыре золотых сигнала'
ON CONFLICT (lesson_id, chunk_index) DO UPDATE SET body_text = EXCLUDED.body_text;

-- Урок 6: Alerting на burn rate (Глава 7 Алертинг)
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, v.chunk_index, v.body_text
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'monitoring-alerting'
JOIN (VALUES
  (1, 'Алертинг — мост между мониторингом и действием. Плохой алертинг хуже отсутствия: 1500 SMS в день, ложные срабатывания — дежурный выгорает и игнорирует оповещения. Принцип: алерт на SLO (исчерпание error budget), а не на симптомы. CPU 90% или latency 2s на 5 минут — не повод будить дежурного, если SLO выполняется. Error budget интегрирует все симптомы в одну метрику. Page (критичный) — будит дежурного, burn rate 14.4x, окно 1ч/5м. Ticket — задача в трекере, burn rate ~1x, окна 3d/6h. Маршрутизация: page — on-call (звонок, push), ticket — отдельный канал (Jira, Slack). Иначе alert fatigue. Два окна в условии алерта: длинное ловит устойчивую деградацию, короткое подтверждает актуальность. Sloth генерирует recording rules и алерты по YAML-спецификации SLO.'),
  (2, 'Требования к алертам: Actionable (есть runbook), Routable (приходит тому, кто может починить), Proportional (severity соответствует влиянию), Confirmed (подтверждение получения). В алерте: что сломалось, насколько серьёзно, дашборд, runbook, дежурный. Соотношение сигнал/шум: на 10 алертов 8 должны быть реальными. Каскад эскалации: Push → не подтверждён 2 мин → Telegram → 3 мин → SMS → 5 мин → звонок → эскалация на второго дежурного. Мониторинг самой системы оповещения обязателен. Анти-паттерны: DDoS на дежурного (тысячи алертов), silence на 60 месяцев вместо 60 минут, алерт без runbook, алерты на симптомы вместо SLO. Runbook хранить в git рядом с кодом.')
) AS v(chunk_index, body_text) ON true
WHERE l.title = 'Alerting на burn rate'
ON CONFLICT (lesson_id, chunk_index) DO UPDATE SET body_text = EXCLUDED.body_text;

-- Урок 7: Инцидентный процесс (Глава 8 Инцидент-менеджмент)
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, v.chunk_index, v.body_text
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'incident-response'
JOIN (VALUES
  (1, 'Жизненный цикл инцидента: Обнаружен → Устраняется → Устранён → Разбор завершён → Все работы завершены. Инцидент закрыт только когда выполнены корректирующие действия. Роли: Incident Commander (IC) — координирует, не чинит руками, отвечает «что происходит и что делаем». Operations Lead — распределяет задачи между инженерами, следит за проверкой гипотез. Communications Lead — апдейты для бизнеса и клиентов, статусная страница. В первые 5 минут: дежурный — временный IC; при эскалации явно передать роль: «Я передаю роль IC тебе, подтверди.» Назначить Operations Lead и Communications Lead. Вести debug doc с самого начала.'),
  (2, 'Debug doc — живой документ во время инцидента: влияние, таймлайн, гипотезы (проверено/не проверено), что осталось для постмортема. Фиксирует факты в момент возникновения. Триггер vs причина: триггер — что непосредственно запустило (релиз, нагрузка); причина — что сделало инцидент возможным (недостаточно тестов, нет fallback). Классификация по тому, кто будет исправлять: баг — разработчики, процессы — руководители, тестирование — QA. Сокращение времени диагностики: метрики зависимостей, плейбуки на алерты, диагностика через очереди (где копится — там узкое место), debug doc. Плановые работы: план с конкретными командами, план отката проверен, точки невозврата, критерии перехода в инцидент. Заведение инцидента — 30 секунд (бот в мессенджере), не 15 полей в Jira.')
) AS v(chunk_index, body_text) ON true
WHERE l.title = 'Инцидентный процесс'
ON CONFLICT (lesson_id, chunk_index) DO UPDATE SET body_text = EXCLUDED.body_text;

-- Урок 8: Blameless postmortem (Глава 9 Постмортемы)
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, v.chunk_index, v.body_text
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'incident-response'
JOIN (VALUES
  (1, 'Постмортем без поиска виновных. Цель — системные причины и действия, а не наказание. Содержимое: контекст и impact, таймлайн, факторы, способствующие сбою, корневые причины, action items с владельцами и сроками. Action items обязательны; без них постмортем ритуальный. Топ-3 причины сбоев по статистике: (1) зарелизили с багом, (2) ошиблись с перенастройкой, (3) что-то переполнилось (память, пулы, диск). Классификация: причина известна и принята как риск (класс 0), ошибка в коде/архитектуре (1), в процессах (2), ошибка сотрудника при работах (3). Ищем не единственную «корневую» причину, а несколько причин, на которые можно повлиять. Sharing sessions: раз в 2–4 недели одна команда рассказывает другой о постмортеме — передача знаний и нормализация обсуждения ошибок.'),
  (2, 'Внедрение с нуля: шаблон, первые три постмортема пишет SRE/тимлид, разбор вслух с командой, action items в трекер в тот же день, на ретро проверять статус P1 action items. Через месяц — сколько написано, сколько закрыто. Через три месяца — sharing sessions. Через полгода — мета-анализ: частые причины, повторяющиеся action items. Плохой постмортем лучше никакого; начать с таймлайн + одна причина + один action item.')
) AS v(chunk_index, body_text) ON true
WHERE l.title = 'Blameless postmortem'
ON CONFLICT (lesson_id, chunk_index) DO UPDATE SET body_text = EXCLUDED.body_text;

-- Урок 9: Что считать Toil (Глава 14 Toil и автоматизация — общие тезисы из книги)
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, v.chunk_index, v.body_text
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'toil-automation'
JOIN (VALUES
  (1, 'Toil — ручная, повторяемая, реактивная работа без долговременной ценности. Примеры: ручной перезапуск сервисов, разовые правки в конфигах, однотипные тикеты. Цель SRE — ограничить долю toil (например не более 50% времени) и сокращать его. Признаки: ручная работа, повторяемость, реактивность (на событие, а не по плану), рост линейно с системой, не создаёт устойчивой ценности. Если toil растёт, команда не успевает заниматься инженерными задачами. Классификация: можно автоматизировать, можно удалить (упростить систему), можно делегировать. Приоритет — автоматизация того, что повторяется чаще всего и отнимает больше всего времени. Метрика: доля времени на toil в спринте; цель — снижение.'),
  (2, 'Автоматизация toil освобождает время для проектной работы и снижает ошибки. «Automate this year''s job away» — постоянное смещение рутины в код. Что автоматизировать в приоритете: шаги runbook (перезапуск, откат), проверки до/после деплоя, откат по критериям (SLO). Идемпотентность и тесты на окружении близком к проду уменьшают ошибки. Автоматизация — не «написать скрипт», а изменить процесс: скрипт без документации и мониторинга становится техдолгом.')
) AS v(chunk_index, body_text) ON true
WHERE l.title = 'Что считать Toil'
ON CONFLICT (lesson_id, chunk_index) DO UPDATE SET body_text = EXCLUDED.body_text;

-- Урок 10: Автоматизация как стратегия надежности
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, 1, 'Автоматизация снижает MTTR и когнитивную нагрузку на дежурных. Приоритеты: runbook-шаги, проверки деплоев, откат по SLO. Идемпотентность и тесты на prod-like окружении. Культура: автоматизация как часть определения «готово» — если делаем вручную больше двух раз, планируем автоматизацию. Плейбуки и runbook в git, версионируются вместе с кодом. Ограничение: не автоматизировать то, что часто меняется и нестабильно, пока процесс не устоялся. «Ночью не надо думать, ночью надо идти по плану» — хороший план плановых работ это конкретные команды, все решения приняты днём.'
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'toil-automation'
WHERE l.title = 'Автоматизация как стратегия надежности'
ON CONFLICT (lesson_id, chunk_index) DO UPDATE SET body_text = EXCLUDED.body_text;

-- Урок 11: Capacity planning (из глав 11 и общих тезисов)
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, v.chunk_index, v.body_text
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'capacity-release'
JOIN (VALUES
  (1, 'Capacity planning — прогнозирование нагрузки и ресурсов, чтобы не уйти в дефицит неожиданно. «Что-то переполнилось» — одна из топ-3 причин сбоев: память, воркеры, пул соединений, диск. Шаги: сбор метрик использования и роста, определение узких мест (CPU, память, сеть, лимиты внешних API), тесты на отказоустойчивость и лимиты. Планировать с запасом, пересматривать регулярно. Сезонное повышение нагрузки (Чёрная пятница, конец квартала) — вопрос: проводилось ли capacity planning? Решения: алерты на ёмкость, автоскейлинг, лимиты и квоты. Медиана RPS за «сейчас ±час» и 1/2/3 недели назад даёт устойчивый прогноз нормальной нагрузки для сравнения с текущей.')
) AS v(chunk_index, body_text) ON true
WHERE l.title = 'Capacity planning'
ON CONFLICT (lesson_id, chunk_index) DO UPDATE SET body_text = EXCLUDED.body_text;

-- Урок 12: Надежные релизы
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, v.chunk_index, v.body_text
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'capacity-release'
JOIN (VALUES
  (1, 'Надёжный релиз: постепенный rollout (canary, процент трафика), явные критерии отката (SLO, error rate, latency), наблюдаемость во время релиза. «Зарелизили с багом» — причина номер один сбоев. Канареечный деплой обязателен при бюджете 20–50%. План отката должен быть проверен на практике. Error budget policy: при бюджете < 20% только баг-фиксы и надёжность; при исчерпании — стоп-кран, релизы запрещены кроме улучшающих надёжность. Аннотации деплоев на графиках SLO: «SLI просел в 14:35 — в 14:32 был релиз auth. Совпадение?» Change budget можно привязать к error budget. Автооткат при срабатывании алертов снижает время простоя.')
) AS v(chunk_index, body_text) ON true
WHERE l.title = 'Надежные релизы'
ON CONFLICT (lesson_id, chunk_index) DO UPDATE SET body_text = EXCLUDED.body_text;

-- Урок 13: Надежность в Kubernetes (Глава 12)
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, v.chunk_index, v.body_text
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'kubernetes-reliability'
JOIN (VALUES
  (1, 'Kubernetes и надёжность: readiness (когда отдавать трафик) и liveness (когда перезапустить). CPU limits — спор: CFS throttling даёт непредсказуемые задержки; без limits один под может занять всю ноду. Решение зависит от языка и контекста; при снятии limits — LimitRange в namespace. Memory: limit = request для Guaranteed QoS. Runtime: GOMAXPROCS и GOMEMLIMIT в Go должны учитывать лимиты контейнера. preStop hook с sleep 5 даёт время kube-proxy обновить правила после удаления пода из endpoints. terminationGracePeriodSeconds, перехват SIGTERM, завершение текущих запросов — иначе rolling update теряет запросы. PodDisruptionBudget (minAvailable/maxUnavailable) — обязательно для критичных сервисов. HPA со стабилизацией (scaleDown медленнее scaleUp) против flapping.'),
  (2, 'Мониторинг k8s для SRE: CPU throttling по подам, memory pressure (>90% лимита), количество рестартов подов, pending поды, утилизация нод. Алерт на частые рестарты (crashloop). Anti-affinity и топология для распределения реплик по зонам. Базовый набор: probes, requests/limits, PDB, graceful shutdown, аннотации деплоев на дашбордах.')
) AS v(chunk_index, body_text) ON true
WHERE l.title = 'Надежность в Kubernetes'
ON CONFLICT (lesson_id, chunk_index) DO UPDATE SET body_text = EXCLUDED.body_text;

-- Урок 14: Хаос и нагрузочное тестирование (Глава 13)
INSERT INTO lesson_content_chunks (lesson_id, chunk_index, body_text)
SELECT l.id, v.chunk_index, v.body_text
FROM lessons l
JOIN modules m ON m.id = l.module_id AND m.slug = 'kubernetes-reliability'
JOIN (VALUES
  (1, 'Хаос-инжиниринг — не «сломать всё», а научный метод: гипотеза («сервис переживёт потерю одной ноды»), эксперимент, наблюдение, выводы. Нагрузочное тестирование подтверждает работу при ожидаемой нагрузке; стресс-тестирование находит точку отказа и карту деградации. Стресс как мини-хаос: ступенчатый рост нагрузки (1x → 2x → 3x), фиксировать при каком RPS деградация и какой компонент первым. Запускать днём при полной команде. Никогда не гонять стресс на проде без предупреждения и плана остановки. Тестовый трафик маркировать (X-Load-Test: true) и исключать из SLI. Плановая перезагрузка нод — мини хаос-тест: проверяет probes, PDB, graceful shutdown, перебалансировку.'),
  (2, 'Практика: ступенчатый профиль (ramp-up) в k6 или аналоге; пороги по p99 и error rate. Результаты документировать с лимитами («auth-service 5xx при >5000 RPS»). CronJob еженедельного rolling restart критичных Deployment — гигиена и проверка автоматики. Пробер «крайнего случая»: не только /healthz, а полный сценарий (создать заказ, проверить данные) для валидации бизнес-логики. «Паучье чутьё» — интуицию формализовать в гипотезы и проверять хаос-экспериментами.')
) AS v(chunk_index, body_text) ON true
WHERE l.title = 'Хаос и нагрузочное тестирование'
ON CONFLICT (lesson_id, chunk_index) DO UPDATE SET body_text = EXCLUDED.body_text;
