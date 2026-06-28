ALTER TABLE domains
    ADD COLUMN IF NOT EXISTS setup_record TEXT,
    ADD COLUMN IF NOT EXISTS setup_dkim_record TEXT,
    ADD COLUMN IF NOT EXISTS setup_verified BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE domains DROP CONSTRAINT IF EXISTS check_domain_status;
ALTER TABLE domains ADD CONSTRAINT check_domain_status
    CHECK (status IN ('pending', 'dns_pending', 'active', 'inactive', 'failed'));
