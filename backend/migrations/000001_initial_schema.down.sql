-- Drop triggers
DROP TRIGGER IF EXISTS update_quotas_updated_at ON quotas;
DROP TRIGGER IF EXISTS update_mailboxes_updated_at ON mailboxes;
DROP TRIGGER IF EXISTS update_domains_updated_at ON domains;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS quotas;
DROP TABLE IF EXISTS aliases;
DROP TABLE IF EXISTS mailboxes;
DROP TABLE IF EXISTS domains;
DROP TABLE IF EXISTS organization_users;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS organizations;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
