package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/middleware"
	"github.com/nivas/server/internal/service"
	"github.com/nivas/server/pkg/apperror"
	"github.com/nivas/server/pkg/response"
)

type PropertyHandler struct {
	svc *service.PropertyService
}

func NewPropertyHandler(svc *service.PropertyService) *PropertyHandler {
	return &PropertyHandler{svc: svc}
}

type createPropertyRequest struct {
	Name    string  `json:"name"`
	Address *string `json:"address"`
}

func (h *PropertyHandler) List(c *gin.Context) {
	props, err := h.svc.List(c.Request.Context(), middleware.GetOrganizationID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	if props == nil {
		props = []domain.Property{}
	}
	response.OK(c, props)
}

func (h *PropertyHandler) Create(c *gin.Context) {
	if !middleware.IsStaffOwner(c) {
		response.Error(c, apperror.Forbidden("only the organization owner can manage properties"))
		return
	}
	var req createPropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	prop, err := h.svc.Create(c.Request.Context(), middleware.GetOrganizationID(c), req.Name, req.Address)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, prop)
}

type updatePropertyRequest struct {
	Name    string  `json:"name"`
	Address *string `json:"address"`
}

func (h *PropertyHandler) Update(c *gin.Context) {
	if !middleware.IsStaffOwner(c) {
		response.Error(c, apperror.Forbidden("only the organization owner can manage properties"))
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	var req updatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	prop, err := h.svc.Update(c.Request.Context(), middleware.GetOrganizationID(c), id, req.Name, req.Address)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, prop)
}

type ExportHandler struct {
	svc *service.ExportService
}

func NewExportHandler(svc *service.ExportService) *ExportHandler {
	return &ExportHandler{svc: svc}
}

func (h *ExportHandler) writeExport(c *gin.Context, result *service.ExportResult, err error) {
	if err != nil {
		response.Error(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+result.Filename)
	c.Data(http.StatusOK, result.ContentType, result.Data)
}

func (h *ExportHandler) Payments(c *gin.Context) {
	propertyID, err := middleware.OptionalPropertyID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	format := c.DefaultQuery("format", "csv")
	result, err := h.svc.Payments(c.Request.Context(), middleware.GetOrganizationID(c), propertyID, format)
	h.writeExport(c, result, err)
}

func (h *ExportHandler) Tenants(c *gin.Context) {
	propertyID, err := middleware.OptionalPropertyID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	format := c.DefaultQuery("format", "csv")
	result, err := h.svc.Tenants(c.Request.Context(), middleware.GetOrganizationID(c), propertyID, format)
	h.writeExport(c, result, err)
}

func (h *ExportHandler) Expenses(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	result, err := h.svc.Expenses(c.Request.Context(), middleware.GetOrganizationID(c), format)
	h.writeExport(c, result, err)
}
