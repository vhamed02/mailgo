package application

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net"
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
}

type DNSRecords struct {
	SPF   string `json:"spf"`
	DKIM  string `json:"dkim"`
	DMARC string `json:"dmarc"`
}

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// AddDomain adds a new domain to an organization
func (s *DomainService) AddDomain(ctx context.Context, req AddDomainRequest) (*domain.Domain, error) {
	// Validate domain name
	domainName := strings.ToLower(strings.TrimSpace(req.Name))
	if !domainRegex.MatchString(domainName) {
		return nil, domain.ErrInvalidDomainName
	}

	// Check if domain already exists
	existing, _ := s.domainRepo.GetByName(ctx, domainName)
	if existing != nil {
		return nil, domain.ErrDomainInUse
	}

	// Check quota
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

	// Generate DNS records
	dnsRecords := s.generateDNSRecords(domainName)

	// Create domain
	dom := &domain.Domain{
		ID:             uuid.New(),
		OrganizationID: req.OrganizationID,
		Name:           domainName,
		Status:         domain.DomainStatusPending,
		DNSVerified:    false,
		SPFRecord:      dnsRecords.SPF,
		DKIMRecord:     dnsRecords.DKIM,
		DMARCRecord:    dnsRecords.DMARC,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.domainRepo.Create(ctx, dom); err != nil {
		return nil, err
	}

	// Enqueue DNS verification check
	if err := s.queue.EnqueueDomainVerification(ctx, dom.ID); err != nil {
		fmt.Printf("Failed to enqueue domain verification: %v\n", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, &domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: req.OrganizationID,
		UserID:         &req.UserID,
		Action:         domain.ActionCreate,
		EntityType:     "domain",
		EntityID:       &dom.ID,
		Details: map[string]interface{}{
			"domain_name": domainName,
		},
		CreatedAt: time.Now(),
	})

	return dom, nil
}

// GetDomain retrieves a domain by ID
func (s *DomainService) GetDomain(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*domain.Domain, error) {
	dom, err := s.domainRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if dom.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}

	return dom, nil
}

// ListDomains lists all domains for an organization
func (s *DomainService) ListDomains(ctx context.Context, orgID uuid.UUID) ([]*domain.Domain, error) {
	return s.domainRepo.ListByOrganization(ctx, orgID)
}

// VerifyDomain triggers DNS verification for a domain
func (s *DomainService) VerifyDomain(ctx context.Context, id uuid.UUID, orgID uuid.UUID, userID uuid.UUID) (*domain.Domain, error) {
	dom, err := s.domainRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dom.OrganizationID != orgID {
		return nil, domain.ErrForbidden
	}

	// In a real implementation, this would check DNS records
	// For now, we'll simulate verification
	verified := s.checkDNSRecords(dom)

	if verified {
		now := time.Now()
		dom.DNSVerified = true
		dom.Status = domain.DomainStatusActive
		dom.VerifiedAt = &now
		dom.UpdatedAt = now

		if err := s.domainRepo.Update(ctx, dom); err != nil {
			return nil, err
		}

		// Audit log
		s.auditRepo.Create(ctx, &domain.AuditLog{
			ID:             uuid.New(),
			OrganizationID: orgID,
			UserID:         &userID,
			Action:         "verify",
			EntityType:     "domain",
			EntityID:       &dom.ID,
			Details: map[string]interface{}{
				"domain_name": dom.Name,
			},
			CreatedAt: time.Now(),
		})
	}

	return dom, nil
}

// DeleteDomain removes a domain
func (s *DomainService) DeleteDomain(ctx context.Context, id uuid.UUID, orgID uuid.UUID, userID uuid.UUID) error {
	dom, err := s.domainRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if dom.OrganizationID != orgID {
		return domain.ErrForbidden
	}

	// TODO: Check if domain has active mailboxes
	// In production, prevent deletion if mailboxes exist or cascade delete

	if err := s.domainRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Audit log
	s.auditRepo.Create(ctx, &domain.AuditLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         &userID,
		Action:         domain.ActionDelete,
		EntityType:     "domain",
		EntityID:       &id,
		Details: map[string]interface{}{
			"domain_name": dom.Name,
		},
		CreatedAt: time.Now(),
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

	records := s.generateDNSRecords(dom.Name)
	dom.SPFRecord = records.SPF
	dom.DKIMRecord = records.DKIM
	dom.DMARCRecord = records.DMARC
	dom.DNSVerified = false
	dom.Status = domain.DomainStatusPending
	dom.UpdatedAt = time.Now()

	if err := s.domainRepo.Update(ctx, dom); err != nil {
		return nil, err
	}
	return dom, nil
}
func (s *DomainService) generateDNSRecords(domainName string) DNSRecords {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	var dkimPublicKey string
	if err == nil {
		pubDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err == nil {
			dkimPublicKey = base64.StdEncoding.EncodeToString(pubDER)
		}
	}
	if dkimPublicKey == "" {
		dkimPublicKey = "GENERATION_FAILED"
	}

	return DNSRecords{
		SPF:   "v=spf1 mx include:mailgo.io ~all",
		DKIM:  fmt.Sprintf("v=DKIM1; k=rsa; p=%s", dkimPublicKey),
		DMARC: fmt.Sprintf("v=DMARC1; p=quarantine; rua=mailto:dmarc@%s", domainName),
	}
}

// checkDNSRecords verifies DNS records are properly configured
// by performing real DNS TXT record lookups.
func (s *DomainService) checkDNSRecords(dom *domain.Domain) bool {
	// Check SPF record
	txts, err := net.LookupTXT(dom.Name)
	if err != nil {
		return false
	}
	for _, txt := range txts {
		if strings.HasPrefix(txt, "v=spf1") {
			return true
		}
	}
	return false
}
