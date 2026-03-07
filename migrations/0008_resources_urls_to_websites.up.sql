-- Заменить local:// ссылки на сайты (для баз, где уже применён старый 0002)
UPDATE resources SET url = 'https://sre.google/sre-book/', title = 'Site Reliability Engineering Book - Managing critical state' WHERE url = 'local://SRE_Коллективный_разум.pdf' AND title LIKE '%Kubernetes%';
UPDATE resources SET url = 'https://sre.google/workbook/', title = 'The Site Reliability Workbook - Testing reliability' WHERE url = 'local://SRE_Коллективный_разум.pdf' AND title LIKE '%Хаос%';
UPDATE resources SET url = 'https://sre.google/books/', title = 'SRE: Коллективный разум (книга RU)' WHERE url = 'local://SRE_Коллективный_разум.pdf';
UPDATE resources SET url = 'https://sre.google/workbook/', title = 'The Site Reliability Workbook' WHERE url = 'local://the-site-reliability-workbook-next18.pdf';

UPDATE quiz_questions SET source_url = 'https://sre.google/books/' WHERE source_url = 'local://SRE_Коллективный_разум.pdf';
