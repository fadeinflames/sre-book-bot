-- Удалить только чанки, добавленные в 0005 (все уроки кроме "Что такое SRE и зачем")
DELETE FROM lesson_content_chunks
WHERE lesson_id IN (
  SELECT id FROM lessons
  WHERE title IN (
    'SRE, DevOps, Platform Engineering',
    'Как формулировать SLI',
    'SLO и Error Budget Policy',
    'Четыре золотых сигнала',
    'Alerting на burn rate',
    'Инцидентный процесс',
    'Blameless postmortem',
    'Что считать Toil',
    'Автоматизация как стратегия надежности',
    'Capacity planning',
    'Надежные релизы',
    'Надежность в Kubernetes',
    'Хаос и нагрузочное тестирование'
  )
);
