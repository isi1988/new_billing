-- Миграция для добавления функционала блокировки клиентов и договоров
-- Добавляем поля is_blocked
ALTER TABLE clients ADD COLUMN IF NOT EXISTS is_blocked BOOLEAN DEFAULT FALSE;
ALTER TABLE contracts ADD COLUMN IF NOT EXISTS is_blocked BOOLEAN DEFAULT FALSE;

-- Создаем индексы для быстрого поиска заблокированных записей
CREATE INDEX IF NOT EXISTS idx_clients_is_blocked ON clients(is_blocked);
CREATE INDEX IF NOT EXISTS idx_contracts_is_blocked ON contracts(is_blocked);