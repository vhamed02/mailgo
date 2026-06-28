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
			spf_record, dkim_record, dmarc_record,
			setup_record, setup_dkim_record, setup_verified,
			verified_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`
	_, err := r.db.Exec(ctx, query,
		d.ID, d.OrganizationID, d.Name, d.Status, d.DNSVerified,
		d.SPFRecord, d.DKIMRecord, d.DMARCRecord,
		d.SetupRecord, d.SetupDKIMRecord, d.SetupVerified,
		d.VerifiedAt, d.CreatedAt, d.UpdatedAt,
	)
	return err
}

func (r *DomainRepository) scanDomain(row interface {
	Scan(dest ...any) error
}) (*domain.Domain, error) {
	d := &domain.Domain{}
	var setupRecord, setupDKIMRecord *string
	err := row.Scan(
		&d.ID, &d.OrganizationID, &d.Name, &d.Status, &d.DNSVerified,
		&d.SPFRecord, &d.DKIMRecord, &d.DMARCRecord,
		&setupRecord, &setupDKIMRecord, &d.SetupVerified,
		&d.VerifiedAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if setupRecord != nil {
		d.SetupRecord = *setupRecord
	}
	if setupDKIMRecord != nil {
		d.SetupDKIMRecord = *setupDKIMRecord
	}
	return d, nil
}

func (r *DomainRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Domain, error) {
	query := `
		SELECT id, organization_id, name, status, dns_verified,
		       spf_record, dkim_record, dmarc_record,
		       setup_record, setup_dkim_record, setup_verified,
		       verified_at, created_at, updated_at
		FROM domains WHERE id = $1
	`
	d, err := r.scanDomain(r.db.QueryRow(ctx, query, id))
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
		       spf_record, dkim_record, dmarc_record,
		       setup_record, setup_dkim_record, setup_verified,
		       verified_at, created_at, updated_at
		FROM domains WHERE name = $1
	`
	d, err := r.scanDomain(r.db.QueryRow(ctx, query, name))
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
		       spf_record, dkim_record, dmarc_record,
		       setup_record, setup_dkim_record, setup_verified,
		       verified_at, created_at, updated_at
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
		d, err := r.scanDomain(rows)
		if err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

func (r *DomainRepository) Update(ctx context.Context, d *domain.Domain) error {
	query := `
		UPDATE domains
		SET status = $1, dns_verified = $2,
		    spf_record = $3, dkim_record = $4, dmarc_record = $5,
		    setup_record = $6, setup_dkim_record = $7, setup_verified = $8,
		    verified_at = $9, updated_at = $10
		WHERE id = $11
	`
	_, err := r.db.Exec(ctx, query,
		d.Status, d.DNSVerified,
		d.SPFRecord, d.DKIMRecord, d.DMARCRecord,
		d.SetupRecord, d.SetupDKIMRecord, d.SetupVerified,
		d.VerifiedAt, d.UpdatedAt,
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
