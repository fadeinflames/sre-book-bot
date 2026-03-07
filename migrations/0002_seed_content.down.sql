DELETE FROM checklist_items
WHERE checklist_id IN (
  SELECT id FROM checklists WHERE title IN (
    'SLO Draft Checklist',
    'Alert Hygiene Checklist',
    'Incident Readiness Checklist',
    'Toil Reduction Checklist'
  )
);

DELETE FROM checklist_submissions
WHERE checklist_id IN (
  SELECT id FROM checklists WHERE title IN (
    'SLO Draft Checklist',
    'Alert Hygiene Checklist',
    'Incident Readiness Checklist',
    'Toil Reduction Checklist'
  )
);

DELETE FROM checklists WHERE title IN (
  'SLO Draft Checklist',
  'Alert Hygiene Checklist',
  'Incident Readiness Checklist',
  'Toil Reduction Checklist'
);

DELETE FROM quiz_answers
WHERE question_id IN (
  SELECT id FROM quiz_questions WHERE source_url IN (
    'https://sre.google/books/',
    'https://sre.google/sre-book/embracing-risk/',
    'https://sre.google/sre-book/service-level-objectives/',
    'https://sre.google/workbook/implementing-slos/',
    'https://sre.google/sre-book/monitoring-distributed-systems/',
    'https://sre.google/workbook/alerting-on-slos/',
    'https://sre.google/workbook/postmortem-culture/',
    'https://sre.google/workbook/incident-response/',
    'https://sre.google/sre-book/eliminating-toil/',
    'https://sre.google/sre-book/automation-at-google/',
    'https://sre.google/sre-book/reliable-product-launches/',
    'https://sre.google/sre-book/capacity-planning/',
    'local://SRE_Коллективный_разум.pdf'
  )
);

DELETE FROM review_items
WHERE question_id IN (
  SELECT id FROM quiz_questions WHERE source_url IN (
    'https://sre.google/books/',
    'https://sre.google/sre-book/embracing-risk/',
    'https://sre.google/sre-book/service-level-objectives/',
    'https://sre.google/workbook/implementing-slos/',
    'https://sre.google/sre-book/monitoring-distributed-systems/',
    'https://sre.google/workbook/alerting-on-slos/',
    'https://sre.google/workbook/postmortem-culture/',
    'https://sre.google/workbook/incident-response/',
    'https://sre.google/sre-book/eliminating-toil/',
    'https://sre.google/sre-book/automation-at-google/',
    'https://sre.google/sre-book/reliable-product-launches/',
    'https://sre.google/sre-book/capacity-planning/',
    'local://SRE_Коллективный_разум.pdf'
  )
);

DELETE FROM quiz_questions WHERE source_url IN (
  'https://sre.google/books/',
  'https://sre.google/sre-book/embracing-risk/',
  'https://sre.google/sre-book/service-level-objectives/',
  'https://sre.google/workbook/implementing-slos/',
  'https://sre.google/sre-book/monitoring-distributed-systems/',
  'https://sre.google/workbook/alerting-on-slos/',
  'https://sre.google/workbook/postmortem-culture/',
  'https://sre.google/workbook/incident-response/',
  'https://sre.google/sre-book/eliminating-toil/',
  'https://sre.google/sre-book/automation-at-google/',
  'https://sre.google/sre-book/reliable-product-launches/',
  'https://sre.google/sre-book/capacity-planning/',
  'local://SRE_Коллективный_разум.pdf'
);

DELETE FROM resources WHERE url IN (
  'https://sre.google/books/',
  'local://SRE_Коллективный_разум.pdf',
  'https://sre.google/workbook/how-sre-relates/',
  'local://the-site-reliability-workbook-next18.pdf',
  'https://sre.google/sre-book/service-level-objectives/',
  'https://sre.google/workbook/implementing-slos/',
  'https://sre.google/sre-book/monitoring-distributed-systems/',
  'https://sre.google/workbook/alerting-on-slos/',
  'https://sre.google/workbook/incident-response/',
  'https://sre.google/workbook/postmortem-culture/',
  'https://sre.google/sre-book/eliminating-toil/',
  'https://sre.google/sre-book/automation-at-google/',
  'https://sre.google/sre-book/reliable-product-launches/'
);

DELETE FROM lessons WHERE title IN (
  'Что такое SRE и зачем',
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
);

DELETE FROM modules WHERE slug IN (
  'foundations',
  'sli-slo-sla',
  'monitoring-alerting',
  'incident-response',
  'toil-automation',
  'capacity-release',
  'kubernetes-reliability'
);
