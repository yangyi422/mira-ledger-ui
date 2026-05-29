ALTER TABLE tags ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

WITH ranked AS (
  SELECT
    id,
    row_number() OVER (PARTITION BY group_name ORDER BY id) AS next_sort_order
  FROM tags
)
UPDATE tags
SET sort_order = (
  SELECT next_sort_order
  FROM ranked
  WHERE ranked.id = tags.id
);
