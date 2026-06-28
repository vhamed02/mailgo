package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailgo/backend/internal/domain"
)

type OrganizationRepository struct {
	db *pgxpool.Pool
}

func NewOrganizationRepository(db *pgxpool.Pool) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(ctx context.Context, org *domain.Organization) error {
	query := `INSERT INTO organizations (id, name, slug, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, org.ID, org.Name, org.Slug, org.Status, org.CreatedAt, org.UpdatedAt)
	return err
}

func (r *OrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	query := `SELECT id, name, slug, status, created_at, updated_at, deleted_at FROM organizations WHERE id = $1`
	org := &domain.Organization{}
	err := r.db.QueryRow(ctx, query, id).Scan(&org.ID, &org.Name, &org.Slug, &org.Status, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return org, err
}

func (r *OrganizationRepository) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	query := `SELECT id, name, slug, status, created_at, updated_at, deleted_at FROM organizations WHERE slug = $1`
	org := &domain.Organization{}
	err := r.db.QueryRow(ctx, query, slug).Scan(&org.ID, &org.Name, &org.Slug, &org.Status, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return org, err
}

func (r *OrganizationRepository) List(ctx context.Context, userID uuid.UUID) ([]*domain.Organization, error) {
	query := `
		SELECT o.id, o.name, o.slug, o.status, o.created_at, o.updated_at, o.deleted_at
		FROM organizations o
		INNER JOIN organization_users ou ON o.id = ou.organization_id
		WHERE ou.user_id = $1 AND o.status != 'deleted'
		ORDER BY o.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*domain.Organization
	for rows.Next() {
		org := &domain.Organization{}
		if err := rows.Scan(&org.ID, &org.Name, &org.Slug, &org.Status, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, rows.Err()
}

func (r *OrganizationRepository) Update(ctx context.Context, org *domain.Organization) error {
	query := `UPDATE organizations SET name = $1, slug = $2, status = $3, updated_at = $4 WHERE id = $5`
	_, err := r.db.Exec(ctx, query, org.Name, org.Slug, org.Status, org.UpdatedAt, org.ID)
	return err
}

func (r *OrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM organizations WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
