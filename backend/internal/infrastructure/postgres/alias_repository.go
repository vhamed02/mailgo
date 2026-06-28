package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailgo/backend/internal/domain"
)

type AliasRepository struct {
	db *pgxpool.Pool
}

func NewAliasRepository(db *pgxpool.Pool) *AliasRepository {
	return &AliasRepository{db: db}
}

func (r *AliasRepository) Create(ctx context.Context, alias *domain.Alias) error {
	query := `
		INSERT INTO aliases (id, organization_id, domain_id, source, destination, active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		alias.ID, alias.OrganizationID, alias.DomainID,
		alias.Source, alias.Destination, alias.Active, alias.CreatedAt,
	)
	return err
}

func (r *AliasRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Alias, error) {
	query := `
		SELECT id, organization_id, domain_id, source, destination, active, created_at
		FROM aliases WHERE id = $1
	`
	a := &domain.Alias{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.OrganizationID, &a.DomainID, &a.Source, &a.Destination, &a.Active, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return a, nil
}

func (r *AliasRepository) GetBySource(ctx context.Context, source string) (*domain.Alias, error) {
	query := `
		SELECT id, organization_id, domain_id, source, destination, active, created_at
		FROM aliases WHERE source = $1
	`
	a := &domain.Alias{}
	err := r.db.QueryRow(ctx, query, source).Scan(
		&a.ID, &a.OrganizationID, &a.DomainID, &a.Source, &a.Destination, &a.Active, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return a, nil
}

func (r *AliasRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*domain.Alias, error) {
	query := `
		SELECT id, organization_id, domain_id, source, destination, active, created_at
		FROM aliases WHERE organization_id = $1
		ORDER BY created_at DESC
	`
	return r.scanAliases(ctx, query, orgID)
}

func (r *AliasRepository) ListByDomain(ctx context.Context, domainID uuid.UUID) ([]*domain.Alias, error) {
	query := `
		SELECT id, organization_id, domain_id, source, destination, active, created_at
		FROM aliases WHERE domain_id = $1
		ORDER BY created_at DESC
	`
	return r.scanAliases(ctx, query, domainID)
}

func (r *AliasRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM aliases WHERE id = $1`, id)
	return err
}

func (r *AliasRepository) CountByOrganization(ctx context.Context, orgID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM aliases WHERE organization_id = $1`,
		orgID,
	).Scan(&count)
	return count, err
}

func (r *AliasRepository) scanAliases(ctx context.Context, query string, arg interface{}) ([]*domain.Alias, error) {
	rows, err := r.db.Query(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aliases []*domain.Alias
	for rows.Next() {
		a := &domain.Alias{}
		if err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.DomainID, &a.Source, &a.Destination, &a.Active, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		aliases = append(aliases, a)
	}
	return aliases, rows.Err()
}
