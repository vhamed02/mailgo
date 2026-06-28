package application

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mailgo/backend/internal/domain"
)

type DomainService struct {
	domainRepo domain.DomainRepository
	quotaRepo  domain.QuotaRepository
	auditRepo  domain.AuditLogRepository
	queue      domain.QueueAdapter
}

func NewDomainService(
	domainRepo domain.DomainRepository,
	quotaRepo domain.QuotaRepository,
	auditRepo domain.AuditLogRepository,
	queue domain.QueueAdapter,
) *DomainService {
	return &DomainService{
		domainRepo: domainRepo,
		quotaRepo:  quotaRepo,
		auditRepo:  auditRepo,
		queue:      queue,
	}
}

type AddDomainRequest struct {
	OrganizationID uuid.UUID
	Name           string
	UserID         uuid.UUID
	UserEmail      string
}

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

func (s *DomainService) AddDomain(ctx context.Context, req AddDomainRequest) (*domain.Domain, error) {
	domainName := strings.ToLower(strings.TrimSpace(req.Name))
	if !domainRegex.MatchString(domainName) {
		return nil, domain.ErrInvalidDomainName
	}

	existing, _ := s.domainRepo.GetByName(ctx, domainName)
	if existing != nil {
		return nil, domain.ErrDomainInUse
	}

	quota, err := s.quotaRepo.GetByOrganization(ctx, req.OrganizationID)
	if err != nil {
		if err == domain.ErrNotFound {
			if createErr := s.quotaRepo.CreateDefault(ctx, req.OrganizationID); createErr != nil {
				return nil, createErr
			}
			quota, err = s.quotaRepo.GetByOrganization(ctx, req.OrganizationID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	currentCount, err := s.domainRepo.CountByOrganization(ctx, req.OrganizationID)
	if err != nil {
		return nil, err
	}
	if currentCount >= quota.MaxDomains {
		return nil, domain.ErrDomainLimitReached
	}

	spf, dkim := s.generateMailcowRecords(domainName)

	dom := &domain.Domain{
		ID:             uuid.New(),
		OrganizationID: req.OrganizationID,
		Name:           domainName,
		Status:         domain.DomainStatusPending,
		DNSVerified:    false,
		SPFRecord:      spf,
		DKIMRecord:     dkim,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.domainRepo.Create(ctx, dom); err != nil {
		return nil, err
	}

	_ = s.queue.EnqueueDomainSetup(ctx, dom.ID)

	s.auditRepo.Create(ctx, &domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: req.OrganizationID,
		UserID:         &req.UserID,
		Action:         domain.ActionCreate,
		EntityType:     "domain",
		EntityID:       &dom.ID,
		Details:        map[string]interface{}{"domain_name": domainName},
		CreatedAt:      time.Now(),
	})

	return dom, nil
}

func (s *DomainService) GetDomain(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.Domain, error) {
	dom, err := s.domainRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dom.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}
	return dom, nil
}

func (s *DomainService) ListDomains(ctx context.Context, orgID uuid.UUID) ([]*domain.Domain, error) {
	return s.domainRepo.ListByOrganization(ctx, orgID)
}

func (s *DomainService) VerifyDomain(ctx context.Context, id uuid.UUID, orgID uuid.UUID, userID uuid.UUID, userEmail string) (*domain.Domain, error) {
	dom, err := s.domainRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dom.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}
	_ = s.queue.EnqueueDomainVerification(ctx, id, userEmail)
	return dom, nil
}

func (s *DomainService) DeleteDomain(ctx context.Context, id uuid.UUID, orgID uuid.UUID, userID uuid.UUID) error {
	dom, err := s.domainRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if dom.OrganizationID != orgID {
		return domain.ErrForbidden
	}
	if err := s.domainRepo.Delete(ctx, id); err != nil {
		return err
	}
	s.auditRepo.Create(ctx, &domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         &userID,
		Action:         domain.ActionDelete,
		EntityType:     "domain",
		EntityID:       &id,
		Details:        map[string]interface{}{"domain_name": dom.Name},
		CreatedAt:      time.Now(),
	})
	return nil
}

func (s *DomainService) RegenerateDNSRecords(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.Domain, error) {
	dom, err := s.domainRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dom.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}
	spf, dkim := s.generateMailcowRecords(dom.Name)
	dom.SPFRecord = spf
	dom.DKIMRecord = dkim
	dom.DNSVerified = false
	dom.BrevoVerified = false
	dom.BrevoAuthenticated = false
	dom.Status = domain.DomainStatusPending
	dom.UpdatedAt = time.Now()
	if err := s.domainRepo.Update(ctx, dom); err != nil {
		return nil, err
	}
	_ = s.queue.EnqueueDomainSetup(ctx, dom.ID)
	return dom, nil
}

func (s *DomainService) generateMailcowRecords(domainName string) (spf, dkim string) {
	spf = "v=spf1 mx include:mailgo.io ~all"

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err == nil {
		pubDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err == nil {
			dkim = fmt.Sprintf("v=DKIM1; k=rsa; p=%s", base64.StdEncoding.EncodeToString(pubDER))
			return
		}
	}
	dkim = "GENERATION_FAILED"
	return
}
