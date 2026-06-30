ALTER TABLE mailboxes
ADD COLUMN IF NOT EXISTS password_encrypted TEXT;
