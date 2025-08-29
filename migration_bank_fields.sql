-- Миграция для добавления раздельных банковских реквизитов
-- Добавляем новые колонки
ALTER TABLE clients ADD COLUMN IF NOT EXISTS bank_name VARCHAR(255);
ALTER TABLE clients ADD COLUMN IF NOT EXISTS bank_account VARCHAR(30);
ALTER TABLE clients ADD COLUMN IF NOT EXISTS bank_bik VARCHAR(9);
ALTER TABLE clients ADD COLUMN IF NOT EXISTS bank_correspondent VARCHAR(30);

-- Миграция данных из старого поля bank_details (если есть)
-- Вы можете вручную перенести данные из bank_details в новые поля

-- После миграции данных можно удалить старое поле:
-- ALTER TABLE clients DROP COLUMN IF EXISTS bank_details;