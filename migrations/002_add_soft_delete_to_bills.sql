-- 002_add_soft_delete_to_bills.sql
-- 为 bills 表增加软删除字段

ALTER TABLE bills ADD COLUMN is_deleted INTEGER NOT NULL DEFAULT 0;
ALTER TABLE bills ADD COLUMN deleted_at TEXT;
