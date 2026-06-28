ALTER TABLE domains DROP COLUMN IF EXISTS setup_record;
ALTER TABLE domains DROP COLUMN IF EXISTS setup_dkim_record;
ALTER TABLE domains DROP COLUMN IF EXISTS setup_verified;

ALTER TABLE domains DROP CONSTRAINT IF EXISTS check_domain_status;
ALTER TABLE domains ADD CONSTRAINT check_domain_status
    CHECK (status IN ('pending', 'active', 'inactive'));
