-- Данные удалённых аккаунтов, которые раньше оставались без владельца
DELETE FROM compliments AS c
WHERE NOT EXISTS (SELECT 1 FROM user_compliment AS uc WHERE uc.compliment_id = c.id);

DELETE FROM important_dates
WHERE user_id IS NULL
  AND partner_id IS NULL;
