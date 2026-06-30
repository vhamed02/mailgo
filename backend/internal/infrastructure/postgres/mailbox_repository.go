package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailgo/backend/internal/domain"
)

type MailboxRepository struct {
	db *pgxpool.Pool
}

func NewMailboxRepository(db *pgxpool.Pool) *MailboxRepository {
	return &MailboxRepository{db: db}
}

func (r *MailboxRepository) Create(ctx context.Context, mailbox *domain.Mailbox) error {
	query := `
		INSERT INTO mailboxes (
			id, organization_id, domain_id, email, local_part,
			display_name, status, quota_bytes, used_bytes, password_hash,
			password_encrypted, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.Exec(ctx, query,
		mailbox.ID,
		mailbox.OrganizationID,
		mailbox.DomainID,
		mailbox.Email,
		mailbox.LocalPart,
		mailbox.DisplayName,
		mailbox.Status,
		mailbox.QuotaBytes,
		mailbox.UsedBytes,
		mailbox.PasswordHash,
		mailbox.PasswordEncrypted,
		mailbox.CreatedAt,
		mailbox.UpdatedAt,
	)

	return err
}

func (r *MailboxRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Mailbox, error) {
	query := `
		SELECT id, organization_id, domain_id, email, local_part,
		       display_name, status, quota_bytes, used_bytes, password_hash,
		       password_encrypted, created_at, updated_at, suspended_at
		FROM mailboxes
		WHERE id = $1
	`

	mailbox := &domain.Mailbox{}
		err := r.db.QueryRow(ctx, query, id).Scan(
		&mailbox.ID,
		&mailbox.OrganizationID,
		&mailbox.DomainID,
		&mailbox.Email,
		&mailbox.LocalPart,
		&mailbox.DisplayName,
		&mailbox.Status,
		&mailbox.QuotaBytes,
		&mailbox.UsedBytes,
		&mailbox.PasswordHash,
		&mailbox.PasswordEncrypted,
		&mailbox.CreatedAt,
		&mailbox.UpdatedAt,
		&mailbox.SuspendedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return mailbox, nil
}

func (r *MailboxRepository) GetByEmail(ctx context.Context, email string) (*domain.Mailbox, error) {
	query := `
		SELECT id, organization_id, domain_id, email, local_part,
		       display_name, status, quota_bytes, used_bytes, password_hash,
		       password_encrypted, created_at, updated_at, suspended_at
		FROM mailboxes
		WHERE email = $1
	`

	mailbox := &domain.Mailbox{}
		err := r.db.QueryRow(ctx, query, email).Scan(
		&mailbox.ID,
		&mailbox.OrganizationID,
		&mailbox.DomainID,
		&mailbox.Email,
		&mailbox.LocalPart,
		&mailbox.DisplayName,
		&mailbox.Status,
		&mailbox.QuotaBytes,
		&mailbox.UsedBytes,
		&mailbox.PasswordHash,
		&mailbox.PasswordEncrypted,
		&mailbox.CreatedAt,
		&mailbox.UpdatedAt,
		&mailbox.SuspendedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return mailbox, nil
}

func (r *MailboxRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*domain.Mailbox, error) {
	query := `
		SELECT id, organization_id, domain_id, email, local_part,
		       display_name, status, quota_bytes, used_bytes, password_hash,
		       password_encrypted, created_at, updated_at, suspended_at
		FROM mailboxes
		WHERE organization_id = $1 AND status != 'deleted'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mailboxes []*domain.Mailbox
	for rows.Next() {
		mailbox := &domain.Mailbox{}
		err := rows.Scan(
			&mailbox.ID,
			&mailbox.OrganizationID,
			&mailbox.DomainID,
			&mailbox.Email,
			&mailbox.LocalPart,
			&mailbox.DisplayName,
			&mailbox.Status,
			&mailbox.QuotaBytes,
			&mailbox.UsedBytes,
			&mailbox.PasswordHash,
			&mailbox.PasswordEncrypted,
			&mailbox.CreatedAt,
			&mailbox.UpdatedAt,
			&mailbox.SuspendedAt,
		)
		if err != nil {
			return nil, err
		}
		mailboxes = append(mailboxes, mailbox)
	}

	return mailboxes, rows.Err()
}

func (r *MailboxRepository) ListByDomain(ctx context.Context, domainID uuid.UUID) ([]*domain.Mailbox, error) {
	query := `
		SELECT id, organization_id, domain_id, email, local_part,
		       display_name, status, quota_bytes, used_bytes, password_hash,
		       password_encrypted, created_at, updated_at, suspended_at
		FROM mailboxes
		WHERE domain_id = $1 AND status != 'deleted'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, domainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mailboxes []*domain.Mailbox
	for rows.Next() {
		mailbox := &domain.Mailbox{}
		err := rows.Scan(
			&mailbox.ID,
			&mailbox.OrganizationID,
			&mailbox.DomainID,
			&mailbox.Email,
			&mailbox.LocalPart,
			&mailbox.DisplayName,
			&mailbox.Status,
			&mailbox.QuotaBytes,
			&mailbox.UsedBytes,
			&mailbox.PasswordHash,
			&mailbox.PasswordEncrypted,
			&mailbox.CreatedAt,
			&mailbox.UpdatedAt,
			&mailbox.SuspendedAt,
		)
		if err != nil {
			return nil, err
		}
		mailboxes = append(mailboxes, mailbox)
	}

	return mailboxes, rows.Err()
}

func (r *MailboxRepository) Update(ctx context.Context, mailbox *domain.Mailbox) error {
	query := `
		UPDATE mailboxes
		SET display_name = $1, status = $2, quota_bytes = $3, used_bytes = $4,
		    password_hash = $5, password_encrypted = $6, suspended_at = $7, updated_at = $8
		WHERE id = $9
	`

	_, err := r.db.Exec(ctx, query,
		mailbox.DisplayName,
		mailbox.Status,
		mailbox.QuotaBytes,
		mailbox.UsedBytes,
		mailbox.PasswordHash,
		mailbox.PasswordEncrypted,
		mailbox.SuspendedAt,
		mailbox.UpdatedAt,
		mailbox.ID,
	)

	return err
}

func (r *MailboxRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM mailboxes WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *MailboxRepository) CountByOrganization(ctx context.Context, orgID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM mailboxes WHERE organization_id = $1 AND status != 'deleted'`
	
	var count int
	err := r.db.QueryRow(ctx, query, orgID).Scan(&count)
	return count, err
}
