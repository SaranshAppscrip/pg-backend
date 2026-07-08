package service

import (
	"context"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/config"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/internal/storage"
	"github.com/nivas/server/pkg/apperror"
)

type DocumentService struct {
	repos   repository.Store
	blobs   storage.BlobStore
	cfg     config.StorageConfig
	presign bool
	ttl     time.Duration
}

func NewDocumentService(repos repository.Store, blobs storage.BlobStore, cfg config.StorageConfig) *DocumentService {
	return &DocumentService{
		repos:   repos,
		blobs:   blobs,
		cfg:     cfg,
		presign: storage.IsPresigned(cfg),
		ttl:     storage.DownloadTTL(cfg),
	}
}

type UploadInput struct {
	OriginalFilename string
	ContentType      string
	SizeBytes        int64
	Body             io.Reader
	Title            string
	ExpiresAt        *time.Time
}

func (s *DocumentService) ListTenantDocuments(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.TenantDocument, error) {
	if _, err := s.repos.Tenants.GetByID(ctx, orgID, tenantID); err != nil {
		return nil, err
	}
	docs, err := s.repos.Documents.ListTenantDocuments(ctx, orgID, tenantID)
	if err != nil {
		return nil, err
	}
	if docs == nil {
		docs = []domain.TenantDocument{}
	}
	return docs, nil
}

func (s *DocumentService) UploadTenantDocument(
	ctx context.Context, orgID, tenantID, staffID uuid.UUID,
	docType domain.TenantDocumentType, in UploadInput,
) (*domain.TenantDocument, error) {
	if _, err := s.repos.Tenants.GetByID(ctx, orgID, tenantID); err != nil {
		return nil, err
	}
	if err := validateUpload(&in, s.cfg.MaxUploadBytes); err != nil {
		return nil, err
	}
	if !isValidTenantDocType(docType) {
		return nil, apperror.BadRequest("invalid document_type")
	}

	docID := uuid.New()
	filename := sanitizeStoredFilename(in.OriginalFilename)
	key := storage.Key(orgID.String(), "tenants", tenantID.String(), docID.String(), filename)

	if err := s.blobs.Put(ctx, key, in.Body, in.SizeBytes, in.ContentType); err != nil {
		return nil, apperror.Internal("store document", err)
	}

	var title *string
	if t := strings.TrimSpace(in.Title); t != "" {
		title = &t
	}
	doc := &domain.TenantDocument{
		ID:               docID,
		OrganizationID:   orgID,
		TenantID:         tenantID,
		DocumentType:     docType,
		Title:            title,
		OriginalFilename: in.OriginalFilename,
		ContentType:      in.ContentType,
		SizeBytes:        in.SizeBytes,
		UploadedBy:       &staffID,
		ExpiresAt:        in.ExpiresAt,
	}
	if err := s.repos.Documents.CreateTenantDocument(ctx, doc, key); err != nil {
		_ = s.blobs.Delete(ctx, key)
		return nil, err
	}
	return doc, nil
}

func (s *DocumentService) TenantDocumentDownload(ctx context.Context, orgID, docID uuid.UUID) (*domain.TenantDocument, string, io.ReadCloser, error) {
	doc, key, err := s.repos.Documents.GetTenantDocument(ctx, orgID, docID)
	if err != nil {
		return nil, "", nil, err
	}
	if s.presign {
		url, err := s.blobs.PresignGet(ctx, key, s.ttl)
		if err != nil {
			return nil, "", nil, apperror.Internal("presign document", err)
		}
		return doc, url, nil, nil
	}
	rc, err := s.blobs.Open(ctx, key)
	if err != nil {
		return nil, "", nil, apperror.Internal("open document", err)
	}
	return doc, "", rc, nil
}

func (s *DocumentService) DeleteTenantDocument(ctx context.Context, orgID, docID uuid.UUID) error {
	doc, key, err := s.repos.Documents.SoftDeleteTenantDocument(ctx, orgID, docID)
	if err != nil {
		return err
	}
	_ = doc
	if err := s.blobs.Delete(ctx, key); err != nil {
		return apperror.Internal("delete document file", err)
	}
	return nil
}

func (s *DocumentService) ListOrganizationDocuments(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID) ([]domain.OrganizationDocument, error) {
	if propertyID != nil {
		if _, err := s.repos.Properties.GetByID(ctx, orgID, *propertyID); err != nil {
			return nil, err
		}
	}
	docs, err := s.repos.Documents.ListOrganizationDocuments(ctx, orgID, propertyID)
	if err != nil {
		return nil, err
	}
	if docs == nil {
		docs = []domain.OrganizationDocument{}
	}
	return docs, nil
}

func (s *DocumentService) UploadOrganizationDocument(
	ctx context.Context, orgID, staffID uuid.UUID, propertyID *uuid.UUID,
	docType domain.OrganizationDocumentType, in UploadInput,
) (*domain.OrganizationDocument, error) {
	if propertyID != nil {
		if _, err := s.repos.Properties.GetByID(ctx, orgID, *propertyID); err != nil {
			return nil, err
		}
	}
	if err := validateUpload(&in, s.cfg.MaxUploadBytes); err != nil {
		return nil, err
	}
	if !isValidOrgDocType(docType) {
		return nil, apperror.BadRequest("invalid document_type")
	}

	docID := uuid.New()
	ownerKey := orgID.String()
	if propertyID != nil {
		ownerKey = propertyID.String()
	}
	filename := sanitizeStoredFilename(in.OriginalFilename)
	key := storage.Key(orgID.String(), "organization", ownerKey, docID.String(), filename)

	if err := s.blobs.Put(ctx, key, in.Body, in.SizeBytes, in.ContentType); err != nil {
		return nil, apperror.Internal("store document", err)
	}

	var title *string
	if t := strings.TrimSpace(in.Title); t != "" {
		title = &t
	}
	doc := &domain.OrganizationDocument{
		ID:               docID,
		OrganizationID:   orgID,
		PropertyID:       propertyID,
		DocumentType:     docType,
		Title:            title,
		OriginalFilename: in.OriginalFilename,
		ContentType:      in.ContentType,
		SizeBytes:        in.SizeBytes,
		UploadedBy:       &staffID,
		ExpiresAt:        in.ExpiresAt,
	}
	if err := s.repos.Documents.CreateOrganizationDocument(ctx, doc, key); err != nil {
		_ = s.blobs.Delete(ctx, key)
		return nil, err
	}
	return doc, nil
}

func (s *DocumentService) OrganizationDocumentDownload(ctx context.Context, orgID, docID uuid.UUID) (*domain.OrganizationDocument, string, io.ReadCloser, error) {
	doc, key, err := s.repos.Documents.GetOrganizationDocument(ctx, orgID, docID)
	if err != nil {
		return nil, "", nil, err
	}
	if s.presign {
		url, err := s.blobs.PresignGet(ctx, key, s.ttl)
		if err != nil {
			return nil, "", nil, apperror.Internal("presign document", err)
		}
		return doc, url, nil, nil
	}
	rc, err := s.blobs.Open(ctx, key)
	if err != nil {
		return nil, "", nil, apperror.Internal("open document", err)
	}
	return doc, "", rc, nil
}

func (s *DocumentService) DeleteOrganizationDocument(ctx context.Context, orgID, docID uuid.UUID) error {
	_, key, err := s.repos.Documents.SoftDeleteOrganizationDocument(ctx, orgID, docID)
	if err != nil {
		return err
	}
	if err := s.blobs.Delete(ctx, key); err != nil {
		return apperror.Internal("delete document file", err)
	}
	return nil
}

func validateUpload(in *UploadInput, maxBytes int64) error {
	if in.Body == nil || in.SizeBytes <= 0 {
		return apperror.BadRequest("file is required")
	}
	if maxBytes > 0 && in.SizeBytes > maxBytes {
		return apperror.BadRequest("file exceeds maximum upload size")
	}
	ct := in.ContentType
	if ct == "" {
		ct = mime.TypeByExtension(filepath.Ext(in.OriginalFilename))
	}
	if ct == "" {
		ct = "application/octet-stream"
	}
	in.ContentType = ct
	if err := storage.ValidateContentType(ct); err != nil {
		return apperror.BadRequest(err.Error())
	}
	return nil
}

func sanitizeStoredFilename(name string) string {
	base := filepath.Base(name)
	if base == "" || base == "." {
		return "document"
	}
	return base
}

func isValidTenantDocType(t domain.TenantDocumentType) bool {
	switch t {
	case domain.TenantDocIDProof, domain.TenantDocLeaseAgreement, domain.TenantDocPoliceVerification,
		domain.TenantDocPhoto, domain.TenantDocOther:
		return true
	}
	return false
}

func isValidOrgDocType(t domain.OrganizationDocumentType) bool {
	switch t {
	case domain.OrgDocPGRegistration, domain.OrgDocFireSafetyNOC, domain.OrgDocPolicePermission,
		domain.OrgDocTradeLicense, domain.OrgDocPropertyTax, domain.OrgDocOther:
		return true
	}
	return false
}
