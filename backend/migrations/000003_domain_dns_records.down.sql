ALTER TABLE domains
    DROP COLUMN IF EXISTS brevo_code_value,
    DROP COLUMN IF EXISTS brevo_dkim1_host,
    DROP COLUMN IF EXISTS brevo_dkim1_value,
    DROP COLUMN IF EXISTS brevo_dkim2_host,
    DROP COLUMN IF EXISTS brevo_dkim2_value,
    DROP COLUMN IF EXISTS brevo_dmarc_value,
    DROP COLUMN IF EXISTS brevo_verified,
    DROP COLUMN IF EXISTS brevo_authenticated,
    ADD COLUMN IF NOT EXISTS setup_record TEXT,
    ADD COLUMN IF NOT EXISTS setup_dkim_record TEXT,
    ADD COLUMN IF NOT EXISTS setup_verified BOOLEAN NOT NULL DEFAULT FALSE;
