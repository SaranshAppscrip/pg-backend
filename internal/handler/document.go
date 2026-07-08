package handler

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/middleware"
	"github.com/nivas/server/internal/service"
	"github.com/nivas/server/pkg/apperror"
	"github.com/nivas/server/pkg/response"
)

type DocumentHandler struct {
	svc *service.DocumentService
}

func NewDocumentHandler(svc *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

func parseExpiresAt(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, apperror.BadRequest("expires_at must be YYYY-MM-DD")
	}
	return &t, nil
}

func readUploadFile(c *gin.Context) (service.UploadInput, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return service.UploadInput{}, apperror.BadRequest("file is required")
	}
	f, err := file.Open()
	if err != nil {
		return service.UploadInput{}, apperror.Internal("open upload", err)
	}
	defer f.Close()

	return service.UploadInput{
		OriginalFilename: file.Filename,
		ContentType:      file.Header.Get("Content-Type"),
		SizeBytes:        file.Size,
		Body:             f,
		Title:            c.PostForm("title"),
	}, nil
}

func (h *DocumentHandler) ListTenantDocuments(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	docs, err := h.svc.ListTenantDocuments(c.Request.Context(), middleware.GetOrganizationID(c), tenantID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, docs)
}

func (h *DocumentHandler) UploadTenantDocument(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	docType := domain.TenantDocumentType(c.PostForm("document_type"))
	in, err := readUploadFile(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	expiresAt, err := parseExpiresAt(c.PostForm("expires_at"))
	if err != nil {
		response.Error(c, err)
		return
	}
	in.ExpiresAt = expiresAt

	doc, err := h.svc.UploadTenantDocument(
		c.Request.Context(),
		middleware.GetOrganizationID(c),
		tenantID,
		middleware.GetUserID(c),
		docType,
		in,
	)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, doc)
}

func (h *DocumentHandler) DownloadTenantDocument(c *gin.Context) {
	h.downloadDocument(c, "tenant")
}

func (h *DocumentHandler) DeleteTenantDocument(c *gin.Context) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteTenantDocument(c.Request.Context(), middleware.GetOrganizationID(c), docID); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

func (h *DocumentHandler) ListOrganizationDocuments(c *gin.Context) {
	propertyID, err := middleware.OptionalPropertyID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	docs, err := h.svc.ListOrganizationDocuments(c.Request.Context(), middleware.GetOrganizationID(c), propertyID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, docs)
}

func (h *DocumentHandler) UploadOrganizationDocument(c *gin.Context) {
	if !middleware.IsStaffOwner(c) {
		response.Error(c, apperror.Forbidden("only the organization owner can upload compliance documents"))
		return
	}
	docType := domain.OrganizationDocumentType(c.PostForm("document_type"))
	in, err := readUploadFile(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	expiresAt, err := parseExpiresAt(c.PostForm("expires_at"))
	if err != nil {
		response.Error(c, err)
		return
	}
	in.ExpiresAt = expiresAt

	var propertyID *uuid.UUID
	if raw := strings.TrimSpace(c.PostForm("property_id")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			response.Error(c, apperror.BadRequest("invalid property_id"))
			return
		}
		propertyID = &id
	}

	doc, err := h.svc.UploadOrganizationDocument(
		c.Request.Context(),
		middleware.GetOrganizationID(c),
		middleware.GetUserID(c),
		propertyID,
		docType,
		in,
	)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, doc)
}

func (h *DocumentHandler) DownloadOrganizationDocument(c *gin.Context) {
	h.downloadDocument(c, "organization")
}

func (h *DocumentHandler) DeleteOrganizationDocument(c *gin.Context) {
	if !middleware.IsStaffOwner(c) {
		response.Error(c, apperror.Forbidden("only the organization owner can delete compliance documents"))
		return
	}
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteOrganizationDocument(c.Request.Context(), middleware.GetOrganizationID(c), docID); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

func (h *DocumentHandler) downloadDocument(c *gin.Context, kind string) {
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	orgID := middleware.GetOrganizationID(c)
	ctx := c.Request.Context()

	var filename, contentType string
	var url string
	var body io.ReadCloser

	switch kind {
	case "tenant":
		doc, presigned, rc, err := h.svc.TenantDocumentDownload(ctx, orgID, docID)
		if err != nil {
			response.Error(c, err)
			return
		}
		filename, contentType, url, body = doc.OriginalFilename, doc.ContentType, presigned, rc
	case "organization":
		doc, presigned, rc, err := h.svc.OrganizationDocumentDownload(ctx, orgID, docID)
		if err != nil {
			response.Error(c, err)
			return
		}
		filename, contentType, url, body = doc.OriginalFilename, doc.ContentType, presigned, rc
	default:
		response.Error(c, apperror.BadRequest("invalid document kind"))
		return
	}

	if url != "" {
		c.Redirect(http.StatusTemporaryRedirect, url)
		return
	}
	defer body.Close()
	c.Header("Content-Disposition", "attachment; filename="+strconv.Quote(filename))
	c.Header("Content-Type", contentType)
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, body)
}
