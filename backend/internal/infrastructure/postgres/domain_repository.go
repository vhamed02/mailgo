package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailgo/backend/internal/domain"
)

type DomainRepository struct {
	db *pgxpool.Pool
}

func NewDomainRepository(db *pgxpool.Pool) *DomainRepository {
	return &DomainRepository{db: db}
}

func (r *DomainRepository) Create(ctx context.Context, d *domain.Domain) error {
	query := `
		INSERT INTO domains (
			id, organization_id, name, status, dns_verified,
			spf_record, dkim_record, dmarc_record, verified_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		d.ID, d.OrganizationID, d.Name, d.Status, d.DNSVerified,
		d.SPFRecord, d.DKIMRecord, d.DMARCRecord, d.VerifiedAt, d.CreatedAt, d.UpdatedAt,
	)
	return err
}

func (r *DomainRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Domain, error) {
	query := `
		SELECT id, organization_id, name, status, dns_verified,
		       spf_record, dkim_record, dmarc_record, verified_at, created_at, updated_at
		FROM domains WHERE id = $1
	`
	d := &domain.Domain{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.OrganizationID, &d.Name, &d.Status, &d.DNSVerified,
		&d.SPFRecord, &d.DKIMRecord, &d.DMARCRecord, &d.VerifiedAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return d, nil
}

func (r *DomainRepository) GetByName(ctx context.Context, name string) (*domain.Domain, error) {
	query := `
		SELECT id, organization_id, name, status, dns_verified,
		       spf_record, dkim_record, dmarc_record, verified_at, created_at, updated_at
		FROM domains WHERE name = $1
	`
	d := &domain.Domain{}
	err := r.db.QueryRow(ctx, query, name).Scan(
		&d.ID, &d.OrganizationID, &d.Name, &d.Status, &d.DNSVerified,
		&d.SPFRecord, &d.DKIMRecord, &d.DMARCRecord, &d.VerifiedAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return d, nil
}

func (r *DomainRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*domain.Domain, error) {
	query := `
		SELECT id, organization_id, name, status, dns_verified,
		       spf_record, dkim_record, dmarc_record, verified_at, created_at, updated_at
		FROM domains
		WHERE organization_id = $1 AND status != 'inactive'
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*domain.Domain
	for rows.Next() {
		d := &domain.Domain{}
		if err := rows.Scan(
			&d.ID, &d.OrganizationID, &d.Name, &d.Status, &d.DNSVerified,
			&d.SPFRecord, &d.DKIMRecord, &d.DMARCRecord, &d.VerifiedAt, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

func (r *DomainRepository) Update(ctx context.Context, d *domain.Domain) error {
	query := `
		UPDATE domains
		SET status = $1, dns_verified = $2, spf_record = $3,
		    dkim_record = $4, dmarc_record = $5, verified_at = $6, updated_at = $7
		WHERE id = $8
	`
	_, err := r.db.Exec(ctx, query,
		d.Status, d.DNSVerified, d.SPFRecord,
		d.DKIMRecord, d.DMARCRecord, d.VerifiedAt, d.UpdatedAt,
		d.ID,
	)
	return err
}

func (r *DomainRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM domains WHERE id = $1`, id)
	return err
}

func (r *DomainRepository) CountByOrganization(ctx context.Context, orgID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM domains WHERE organization_id = $1 AND status != 'inactive'`,
		orgID,
	).Scan(&count)
	return count, err
}
