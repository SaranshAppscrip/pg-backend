package postgres

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/pkg/apperror"
)

func sanitizeFilename(name string) string {
	base := filepath.Base(name)
	base = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == 0 {
			return -1
		}
		return r
	}, base)
	if base == "" || base == "." {
		return "document"
	}
	return base
}

func (s *Store) CreateTenantDocument(ctx context.Context, doc *domain.TenantDocument, storageKey string) error {
	err := s.pool.QueryRow(ctx, `
		INSERT INTO tenant_documents (
			id, organization_id, tenant_id, document_type, title, storage_key,
			original_filename, content_type, size_bytes, uploaded_by, expires_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING created_at
	`, doc.ID, doc.OrganizationID, doc.TenantID, doc.DocumentType, doc.Title, storageKey,
		doc.OriginalFilename, doc.ContentType, doc.SizeBytes, doc.UploadedBy, doc.ExpiresAt,
	).Scan(&doc.CreatedAt)
	return mapPgError(err, "tenant document")
}

func (s *Store) ListTenantDocuments(ctx context.Context, orgID, tenantID uuid.UUID) ([]domain.TenantDocument, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, tenant_id, document_type, title,
		       original_filename, content_type, size_bytes, uploaded_by, expires_at, created_at
		FROM tenant_documents
		WHERE organization_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, orgID, tenantID)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()
	return scanTenantDocuments(rows)
}

func (s *Store) GetTenantDocument(ctx context.Context, orgID, id uuid.UUID) (*domain.TenantDocument, string, error) {
	var doc domain.TenantDocument
	var storageKey string
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, tenant_id, document_type, title, storage_key,
		       original_filename, content_type, size_bytes, uploaded_by, expires_at, created_at
		FROM tenant_documents
		WHERE organization_id = $1 AND id = $2 AND deleted_at IS NULL
	`, orgID, id).Scan(
		&doc.ID, &doc.OrganizationID, &doc.TenantID, &doc.DocumentType, &doc.Title, &storageKey,
		&doc.OriginalFilename, &doc.ContentType, &doc.SizeBytes, &doc.UploadedBy, &doc.ExpiresAt, &doc.CreatedAt,
	)
	if err != nil {
		return nil, "", mapPgError(err, "tenant document")
	}
	return &doc, storageKey, nil
}

func (s *Store) SoftDeleteTenantDocument(ctx context.Context, orgID, id uuid.UUID) (*domain.TenantDocument, string, error) {
	var doc domain.TenantDocument
	var storageKey string
	err := s.pool.QueryRow(ctx, `
		UPDATE tenant_documents SET deleted_at = NOW()
		WHERE organization_id = $1 AND id = $2 AND deleted_at IS NULL
		RETURNING id, organization_id, tenant_id, document_type, title, storage_key,
		          original_filename, content_type, size_bytes, uploaded_by, expires_at, created_at
	`, orgID, id).Scan(
		&doc.ID, &doc.OrganizationID, &doc.TenantID, &doc.DocumentType, &doc.Title, &storageKey,
		&doc.OriginalFilename, &doc.ContentType, &doc.SizeBytes, &doc.UploadedBy, &doc.ExpiresAt, &doc.CreatedAt,
	)
	if err != nil {
		return nil, "", mapPgError(err, "tenant document")
	}
	return &doc, storageKey, nil
}

func (s *Store) CreateOrganizationDocument(ctx context.Context, doc *domain.OrganizationDocument, storageKey string) error {
	err := s.pool.QueryRow(ctx, `
		INSERT INTO organization_documents (
			id, organization_id, property_id, document_type, title, storage_key,
			original_filename, content_type, size_bytes, uploaded_by, expires_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING created_at
	`, doc.ID, doc.OrganizationID, doc.PropertyID, doc.DocumentType, doc.Title, storageKey,
		doc.OriginalFilename, doc.ContentType, doc.SizeBytes, doc.UploadedBy, doc.ExpiresAt,
	).Scan(&doc.CreatedAt)
	return mapPgError(err, "organization document")
}

func (s *Store) ListOrganizationDocuments(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID) ([]domain.OrganizationDocument, error) {
	query := `
		SELECT id, organization_id, property_id, document_type, title,
		       original_filename, content_type, size_bytes, uploaded_by, expires_at, created_at
		FROM organization_documents
		WHERE organization_id = $1 AND deleted_at IS NULL`
	args := []any{orgID}
	if propertyID != nil {
		query += ` AND (property_id IS NULL OR property_id = $2)`
		args = append(args, *propertyID)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()
	return scanOrganizationDocuments(rows)
}

func (s *Store) GetOrganizationDocument(ctx context.Context, orgID, id uuid.UUID) (*domain.OrganizationDocument, string, error) {
	var doc domain.OrganizationDocument
	var storageKey string
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, property_id, document_type, title, storage_key,
		       original_filename, content_type, size_bytes, uploaded_by, expires_at, created_at
		FROM organization_documents
		WHERE organization_id = $1 AND id = $2 AND deleted_at IS NULL
	`, orgID, id).Scan(
		&doc.ID, &doc.OrganizationID, &doc.PropertyID, &doc.DocumentType, &doc.Title, &storageKey,
		&doc.OriginalFilename, &doc.ContentType, &doc.SizeBytes, &doc.UploadedBy, &doc.ExpiresAt, &doc.CreatedAt,
	)
	if err != nil {
		return nil, "", mapPgError(err, "organization document")
	}
	return &doc, storageKey, nil
}

func (s *Store) SoftDeleteOrganizationDocument(ctx context.Context, orgID, id uuid.UUID) (*domain.OrganizationDocument, string, error) {
	var doc domain.OrganizationDocument
	var storageKey string
	err := s.pool.QueryRow(ctx, `
		UPDATE organization_documents SET deleted_at = NOW()
		WHERE organization_id = $1 AND id = $2 AND deleted_at IS NULL
		RETURNING id, organization_id, property_id, document_type, title, storage_key,
		          original_filename, content_type, size_bytes, uploaded_by, expires_at, created_at
	`, orgID, id).Scan(
		&doc.ID, &doc.OrganizationID, &doc.PropertyID, &doc.DocumentType, &doc.Title, &storageKey,
		&doc.OriginalFilename, &doc.ContentType, &doc.SizeBytes, &doc.UploadedBy, &doc.ExpiresAt, &doc.CreatedAt,
	)
	if err != nil {
		return nil, "", mapPgError(err, "organization document")
	}
	return &doc, storageKey, nil
}

func scanTenantDocuments(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]domain.TenantDocument, error) {
	var list []domain.TenantDocument
	for rows.Next() {
		var doc domain.TenantDocument
		if err := rows.Scan(
			&doc.ID, &doc.OrganizationID, &doc.TenantID, &doc.DocumentType, &doc.Title,
			&doc.OriginalFilename, &doc.ContentType, &doc.SizeBytes, &doc.UploadedBy, &doc.ExpiresAt, &doc.CreatedAt,
		); err != nil {
			return nil, apperror.Internal("scan tenant document", err)
		}
		list = append(list, doc)
	}
	return list, rows.Err()
}

func scanOrganizationDocuments(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]domain.OrganizationDocument, error) {
	var list []domain.OrganizationDocument
	for rows.Next() {
		var doc domain.OrganizationDocument
		if err := rows.Scan(
			&doc.ID, &doc.OrganizationID, &doc.PropertyID, &doc.DocumentType, &doc.Title,
			&doc.OriginalFilename, &doc.ContentType, &doc.SizeBytes, &doc.UploadedBy, &doc.ExpiresAt, &doc.CreatedAt,
		); err != nil {
			return nil, apperror.Internal("scan organization document", err)
		}
		list = append(list, doc)
	}
	return list, rows.Err()
}
