DELETE FROM lesson_content_chunks
WHERE lesson_id IN (
  SELECT id FROM lessons l
  JOIN modules m ON m.id = l.module_id AND m.slug = 'foundations'
  WHERE l.title = 'Что такое SRE и зачем'
);
