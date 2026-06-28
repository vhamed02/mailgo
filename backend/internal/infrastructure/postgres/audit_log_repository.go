package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailgo/backend/internal/domain"
)

type AuditLogRepository struct {
	db *pgxpool.Pool
}

func NewAuditLogRepository(db *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, organization_id, user_id, action, entity_type, entity_id, details, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		log.ID, log.OrganizationID, log.UserID, log.Action, log.EntityType,
		log.EntityID, log.Details, log.IPAddress, log.UserAgent, log.CreatedAt,
	)
	return err
}

func (r *AuditLogRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, organization_id, user_id, action, entity_type, entity_id, details, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		log := &domain.AuditLog{}
		if err := rows.Scan(
			&log.ID, &log.OrganizationID, &log.UserID, &log.Action, &log.EntityType,
			&log.EntityID, &log.Details, &log.IPAddress, &log.UserAgent, &log.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (r *AuditLogRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, organization_id, user_id, action, entity_type, entity_id, details, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		log := &domain.AuditLog{}
		if err := rows.Scan(
			&log.ID, &log.OrganizationID, &log.UserID, &log.Action, &log.EntityType,
			&log.EntityID, &log.Details, &log.IPAddress, &log.UserAgent, &log.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}
