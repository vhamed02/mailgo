package postgres

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailgo/backend/internal/domain"
)

// DomainRepository stub
type DomainRepository struct{ db *pgxpool.Pool }
func NewDomainRepository(db *pgxpool.Pool) *DomainRepository { return &DomainRepository{db: db} }
func (r *DomainRepository) Create(ctx context.Context, domain *domain.Domain) error { return nil }
func (r *DomainRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Domain, error) { return nil, nil }
func (r *DomainRepository) GetByName(ctx context.Context, name string) (*domain.Domain, error) { return nil, nil }
func (r *DomainRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*domain.Domain, error) { return nil, nil }
func (r *DomainRepository) Update(ctx context.Context, domain *domain.Domain) error { return nil }
func (r *DomainRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (r *DomainRepository) CountByOrganization(ctx context.Context, orgID uuid.UUID) (int, error) { return 0, nil }

// AliasRepository stub
type AliasRepository struct{ db *pgxpool.Pool }
func NewAliasRepository(db *pgxpool.Pool) *AliasRepository { return &AliasRepository{db: db} }
func (r *AliasRepository) Create(ctx context.Context, alias *domain.Alias) error { return nil }
func (r *AliasRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Alias, error) { return nil, nil }
func (r *AliasRepository) GetBySource(ctx context.Context, source string) (*domain.Alias, error) { return nil, nil }
func (r *AliasRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*domain.Alias, error) { return nil, nil }
func (r *AliasRepository) ListByDomain(ctx context.Context, domainID uuid.UUID) ([]*domain.Alias, error) { return nil, nil }
func (r *AliasRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (r *AliasRepository) CountByOrganization(ctx context.Context, orgID uuid.UUID) (int, error) { return 0, nil }
