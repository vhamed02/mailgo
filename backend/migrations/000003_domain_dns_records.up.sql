ALTER TABLE domains
    DROP COLUMN IF EXISTS setup_record,
    DROP COLUMN IF EXISTS setup_dkim_record,
    DROP COLUMN IF EXISTS setup_verified,
    ADD COLUMN IF NOT EXISTS brevo_code_value TEXT,
    ADD COLUMN IF NOT EXISTS brevo_dkim1_host TEXT,
    ADD COLUMN IF NOT EXISTS brevo_dkim1_value TEXT,
    ADD COLUMN IF NOT EXISTS brevo_dkim2_host TEXT,
    ADD COLUMN IF NOT EXISTS brevo_dkim2_value TEXT,
    ADD COLUMN IF NOT EXISTS brevo_dmarc_value TEXT,
    ADD COLUMN IF NOT EXISTS brevo_verified BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS brevo_authenticated BOOLEAN NOT NULL DEFAULT FALSE;
