package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/nivas/server/internal/middleware"
	"github.com/nivas/server/internal/service"
	"github.com/nivas/server/pkg/apperror"
	"github.com/nivas/server/pkg/response"
)

type ReminderHandler struct {
	svc *service.ReminderService
}

func NewReminderHandler(svc *service.ReminderService) *ReminderHandler {
	return &ReminderHandler{svc: svc}
}

func (h *ReminderHandler) Run(c *gin.Context) {
	if !middleware.IsStaffOwner(c) {
		response.Error(c, apperror.Forbidden("only the organization owner can trigger rent reminders"))
		return
	}
	reminderType := c.DefaultQuery("type", "due")
	if reminderType != "due" && reminderType != "overdue" {
		response.Error(c, apperror.BadRequest("type must be due or overdue"))
		return
	}
	force := c.Query("force") == "true"

	sent, err := h.svc.RunForOrg(c.Request.Context(), middleware.GetOrganizationID(c), reminderType, force)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{
		"sent": sent,
		"type": reminderType,
		"force": force,
	})
}
