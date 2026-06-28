package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailgo/backend/internal/domain"
	"time"
)

type QuotaRepository struct {
	db *pgxpool.Pool
}

func NewQuotaRepository(db *pgxpool.Pool) *QuotaRepository {
	return &QuotaRepository{db: db}
}

func (r *QuotaRepository) GetByOrganization(ctx context.Context, orgID uuid.UUID) (*domain.Quota, error) {
	query := `
		SELECT id, organization_id, max_domains, max_mailboxes, max_storage_gb, max_aliases, updated_at
		FROM quotas
		WHERE organization_id = $1
	`
	quota := &domain.Quota{}
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&quota.ID, &quota.OrganizationID, &quota.MaxDomains, &quota.MaxMailboxes,
		&quota.MaxStorageGB, &quota.MaxAliases, &quota.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return quota, err
}

func (r *QuotaRepository) CreateDefault(ctx context.Context, orgID uuid.UUID) error {
	query := `
		INSERT INTO quotas (id, organization_id, max_domains, max_mailboxes, max_storage_gb, max_aliases, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		uuid.New(),
		orgID,
		5,   // Default: 5 domains
		50,  // Default: 50 mailboxes
		10,  // Default: 10 GB storage
		100, // Default: 100 aliases
		time.Now(),
	)
	return err
}

func (r *QuotaRepository) Update(ctx context.Context, quota *domain.Quota) error {
	query := `
		UPDATE quotas
		SET max_domains = $1, max_mailboxes = $2, max_storage_gb = $3, max_aliases = $4, updated_at = $5
		WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query,
		quota.MaxDomains, quota.MaxMailboxes, quota.MaxStorageGB, quota.MaxAliases,
		time.Now(), quota.ID,
	)
	return err
}
