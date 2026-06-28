package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailgo/backend/internal/domain"
)

type OrganizationUserRepository struct {
	db *pgxpool.Pool
}

func NewOrganizationUserRepository(db *pgxpool.Pool) *OrganizationUserRepository {
	return &OrganizationUserRepository{db: db}
}

func (r *OrganizationUserRepository) Create(ctx context.Context, orgUser *domain.OrganizationUser) error {
	query := `
		INSERT INTO organization_users (id, organization_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		orgUser.ID, orgUser.OrganizationID, orgUser.UserID, orgUser.Role, orgUser.JoinedAt,
	)
	return err
}

func (r *OrganizationUserRepository) GetByOrganizationAndUser(ctx context.Context, orgID, userID uuid.UUID) (*domain.OrganizationUser, error) {
	query := `
		SELECT id, organization_id, user_id, role, joined_at
		FROM organization_users
		WHERE organization_id = $1 AND user_id = $2
	`
	orgUser := &domain.OrganizationUser{}
	err := r.db.QueryRow(ctx, query, orgID, userID).Scan(
		&orgUser.ID, &orgUser.OrganizationID, &orgUser.UserID, &orgUser.Role, &orgUser.JoinedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return orgUser, err
}

func (r *OrganizationUserRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.OrganizationUser, error) {
	query := `
		SELECT id, organization_id, user_id, role, joined_at
		FROM organization_users
		WHERE user_id = $1
		ORDER BY joined_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgUsers []*domain.OrganizationUser
	for rows.Next() {
		orgUser := &domain.OrganizationUser{}
		if err := rows.Scan(&orgUser.ID, &orgUser.OrganizationID, &orgUser.UserID, &orgUser.Role, &orgUser.JoinedAt); err != nil {
			return nil, err
		}
		orgUsers = append(orgUsers, orgUser)
	}
	return orgUsers, rows.Err()
}

func (r *OrganizationUserRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*domain.OrganizationUser, error) {
	query := `
		SELECT id, organization_id, user_id, role, joined_at
		FROM organization_users
		WHERE organization_id = $1
		ORDER BY joined_at ASC
	`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgUsers []*domain.OrganizationUser
	for rows.Next() {
		orgUser := &domain.OrganizationUser{}
		if err := rows.Scan(&orgUser.ID, &orgUser.OrganizationID, &orgUser.UserID, &orgUser.Role, &orgUser.JoinedAt); err != nil {
			return nil, err
		}
		orgUsers = append(orgUsers, orgUser)
	}
	return orgUsers, rows.Err()
}

func (r *OrganizationUserRepository) Update(ctx context.Context, orgUser *domain.OrganizationUser) error {
	query := `
		UPDATE organization_users
		SET role = $1
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, orgUser.Role, orgUser.ID)
	return err
}

func (r *OrganizationUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM organization_users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
