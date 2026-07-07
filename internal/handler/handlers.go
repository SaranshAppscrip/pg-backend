package handler

import (
	"fmt"
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

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type loginRequest struct {
	OrganizationID *string `json:"organization_id"`
	Email          string  `json:"email"`
	Password       string  `json:"password"`
}

func parseOptionalOrgID(raw *string) (*uuid.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	s := strings.TrimSpace(*raw)
	if s == "" {
		return nil, nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, apperror.BadRequest("invalid organization_id")
	}
	return &id, nil
}

func (h *AuthHandler) StaffLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	orgID, err := parseOptionalOrgID(req.OrganizationID)
	if err != nil {
		response.Error(c, err)
		return
	}
	res, err := h.auth.StaffLogin(c.Request.Context(), req.Email, req.Password, orgID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

func (h *AuthHandler) StaffMe(c *gin.Context) {
	claims := middleware.GetClaims(c)
	profile, err := h.auth.StaffMe(c.Request.Context(), claims.OrganizationID, claims.UserID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, profile)
}

func (h *AuthHandler) StaffLogout(c *gin.Context) {
	var req refreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	if err := h.auth.StaffLogout(c.Request.Context(), req.RefreshToken); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) StaffRefresh(c *gin.Context) {
	var req refreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	res, err := h.auth.StaffRefresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

type forgotPasswordRequest struct {
	OrganizationID *string `json:"organization_id"`
	Email          string  `json:"email"`
}

func (h *AuthHandler) StaffForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	orgID, err := parseOptionalOrgID(req.OrganizationID)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.auth.StaffForgotPassword(c.Request.Context(), req.Email, orgID); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"message": "If an account exists, a reset email was sent"})
}

type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (h *AuthHandler) StaffResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	if err := h.auth.StaffResetPassword(c.Request.Context(), req.Token, req.Password); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"message": "Password updated successfully"})
}

func (h *AuthHandler) TenantLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	orgID, err := parseOptionalOrgID(req.OrganizationID)
	if err != nil {
		response.Error(c, err)
		return
	}
	res, err := h.auth.TenantLogin(c.Request.Context(), req.Email, req.Password, orgID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

func (h *AuthHandler) TenantForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	orgID, err := parseOptionalOrgID(req.OrganizationID)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.auth.TenantForgotPassword(c.Request.Context(), req.Email, orgID); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"message": "If an account exists, a reset email was sent"})
}

func (h *AuthHandler) TenantResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	if err := h.auth.TenantResetPassword(c.Request.Context(), req.Token, req.Password); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"message": "Password updated successfully"})
}

func (h *AuthHandler) TenantMe(c *gin.Context) {
	claims := middleware.GetClaims(c)
	profile, err := h.auth.TenantMe(c.Request.Context(), claims.OrganizationID, claims.UserID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, profile)
}

func (h *AuthHandler) TenantLogout(c *gin.Context) {
	var req refreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	if err := h.auth.TenantLogout(c.Request.Context(), req.RefreshToken); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

func (h *AuthHandler) TenantRefresh(c *gin.Context) {
	var req refreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	res, err := h.auth.TenantRefresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

type SettingsHandler struct {
	svc *service.SettingsService
}

func NewSettingsHandler(svc *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{svc: svc}
}

type updateSettingsRequest struct {
	PGName string `json:"pg_name"`
}

func (h *SettingsHandler) Get(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	res, err := h.svc.Get(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

func (h *SettingsHandler) Update(c *gin.Context) {
	var req updateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	orgID := middleware.GetOrganizationID(c)
	res, err := h.svc.Update(c.Request.Context(), orgID, req.PGName)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

type RoomHandler struct {
	svc *service.RoomService
}

func NewRoomHandler(svc *service.RoomService) *RoomHandler {
	return &RoomHandler{svc: svc}
}

type createRoomRequest struct {
	RoomNumber string `json:"room_number"`
	Capacity   int    `json:"capacity"`
}

func (h *RoomHandler) List(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	rooms, err := h.svc.List(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, err)
		return
	}
	if rooms == nil {
		rooms = []domain.Room{}
	}
	response.OK(c, rooms)
}

func (h *RoomHandler) Create(c *gin.Context) {
	var req createRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	room, err := h.svc.Create(c.Request.Context(), middleware.GetOrganizationID(c), req.RoomNumber, req.Capacity)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, room)
}

func (h *RoomHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), middleware.GetOrganizationID(c), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

type TenantHandler struct {
	svc *service.TenantService
}

func NewTenantHandler(svc *service.TenantService) *TenantHandler {
	return &TenantHandler{svc: svc}
}

type createTenantRequest struct {
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Password   string  `json:"password"`
	Phone      string  `json:"phone"`
	RoomID     string  `json:"room_id"`
	MonthlyFee float64 `json:"monthly_fee"`
	JoinDate   string  `json:"join_date"`
}

func (h *TenantHandler) List(c *gin.Context) {
	tenants, err := h.svc.List(c.Request.Context(), middleware.GetOrganizationID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	if tenants == nil {
		tenants = []domain.Tenant{}
	}
	response.OK(c, tenants)
}

func (h *TenantHandler) Create(c *gin.Context) {
	var req createTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	roomID, err := uuid.Parse(req.RoomID)
	if err != nil {
		response.Error(c, err)
		return
	}
	joinDate, err := time.Parse("2006-01-02", req.JoinDate)
	if err != nil {
		response.Error(c, err)
		return
	}
	tenant, err := h.svc.Create(c.Request.Context(), middleware.GetOrganizationID(c), service.CreateTenantInput{
		Name: req.Name, Email: req.Email, Password: req.Password, Phone: req.Phone, RoomID: roomID,
		MonthlyFee: req.MonthlyFee, JoinDate: joinDate,
	})
	if err != nil {
		fmt.Printf("err: %v\n", err)
		response.Error(c, err)
		return
	}
	response.Created(c, tenant)
}

func (h *TenantHandler) MoveOut(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	tenant, err := h.svc.MoveOut(c.Request.Context(), middleware.GetOrganizationID(c), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, tenant)
}

type PaymentHandler struct {
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

type createPaymentRequest struct {
	TenantID string  `json:"tenant_id"`
	Amount   float64 `json:"amount"`
	Date     string  `json:"date"`
	ForMonth string  `json:"for_month"`
	Mode     string  `json:"mode"`
}

func (h *PaymentHandler) List(c *gin.Context) {
	payments, err := h.svc.List(c.Request.Context(), middleware.GetOrganizationID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	if payments == nil {
		payments = []domain.Payment{}
	}
	response.OK(c, payments)
}

func (h *PaymentHandler) Create(c *gin.Context) {
	var req createPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		response.Error(c, err)
		return
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		response.Error(c, err)
		return
	}
	payment, err := h.svc.Create(c.Request.Context(), middleware.GetOrganizationID(c), service.CreatePaymentInput{
		TenantID: tenantID, Amount: req.Amount, Date: date,
		ForMonth: req.ForMonth, Mode: domain.PaymentMode(req.Mode),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, payment)
}

func (h *PaymentHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), middleware.GetOrganizationID(c), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

func (h *PaymentHandler) TenantPayments(c *gin.Context) {
	claims := middleware.GetClaims(c)
	payments, err := h.svc.ListByTenant(c.Request.Context(), claims.OrganizationID, claims.UserID)
	if err != nil {
		response.Error(c, err)
		return
	}
	if payments == nil {
		payments = []domain.Payment{}
	}
	response.OK(c, payments)
}

type ExpenseHandler struct {
	svc *service.ExpenseService
}

func NewExpenseHandler(svc *service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{svc: svc}
}

type createExpenseRequest struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Date     string  `json:"date"`
	Note     *string `json:"note"`
}

func (h *ExpenseHandler) List(c *gin.Context) {
	expenses, err := h.svc.List(c.Request.Context(), middleware.GetOrganizationID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	if expenses == nil {
		expenses = []domain.Expense{}
	}
	response.OK(c, expenses)
}

func (h *ExpenseHandler) Create(c *gin.Context) {
	var req createExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		response.Error(c, err)
		return
	}
	expense, err := h.svc.Create(c.Request.Context(), middleware.GetOrganizationID(c), service.CreateExpenseInput{
		Category: domain.ExpenseCategory(req.Category),
		Amount:   req.Amount, Date: date, Note: req.Note,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, expense)
}

func (h *ExpenseHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), middleware.GetOrganizationID(c), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

type KitchenHandler struct {
	svc *service.KitchenService
}

func NewKitchenHandler(svc *service.KitchenService) *KitchenHandler {
	return &KitchenHandler{svc: svc}
}

type createKitchenItemRequest struct {
	Name             string  `json:"name"`
	Qty              float64 `json:"qty"`
	Unit             string  `json:"unit"`
	ReorderThreshold float64 `json:"reorder_threshold"`
}

type stockMovementRequest struct {
	Qty  float64 `json:"qty"`
	Date string  `json:"date"`
	Note *string `json:"note"`
}

func (h *KitchenHandler) ListItems(c *gin.Context) {
	items, err := h.svc.ListItems(c.Request.Context(), middleware.GetOrganizationID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	if items == nil {
		items = []domain.KitchenItem{}
	}
	response.OK(c, items)
}

func (h *KitchenHandler) CreateItem(c *gin.Context) {
	var req createKitchenItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	item, err := h.svc.CreateItem(c.Request.Context(), middleware.GetOrganizationID(c), service.CreateKitchenItemInput{
		Name: req.Name, Qty: req.Qty,
		Unit: domain.KitchenUnit(req.Unit), ReorderThreshold: req.ReorderThreshold,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, item)
}

func (h *KitchenHandler) StockIn(c *gin.Context) {
	h.stockAction(c, true)
}

func (h *KitchenHandler) UseStock(c *gin.Context) {
	h.stockAction(c, false)
}

func (h *KitchenHandler) stockAction(c *gin.Context, stockIn bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	var req stockMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		response.Error(c, err)
		return
	}
	input := service.StockMovementInput{Qty: req.Qty, Date: date, Note: req.Note}
	orgID := middleware.GetOrganizationID(c)
	var item *domain.KitchenItem
	if stockIn {
		item, err = h.svc.StockIn(c.Request.Context(), orgID, id, input)
	} else {
		item, err = h.svc.UseStock(c.Request.Context(), orgID, id, input)
	}
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

func (h *KitchenHandler) ListLog(c *gin.Context) {
	limit := 20
	logs, err := h.svc.ListLog(c.Request.Context(), middleware.GetOrganizationID(c), limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	if logs == nil {
		logs = []domain.KitchenLog{}
	}
	response.OK(c, logs)
}

type StaffHandler struct {
	svc *service.StaffService
}

func NewStaffHandler(svc *service.StaffService) *StaffHandler {
	return &StaffHandler{svc: svc}
}

func (h *StaffHandler) List(c *gin.Context) {
	staff, err := h.svc.List(c.Request.Context(), middleware.GetOrganizationID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	if staff == nil {
		staff = []domain.StaffProfile{}
	}
	response.OK(c, staff)
}

type inviteStaffRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *StaffHandler) Invite(c *gin.Context) {
	if !middleware.IsStaffOwner(c) {
		response.Error(c, apperror.Forbidden("only the organization owner can invite staff"))
		return
	}
	var req inviteStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}
	profile, err := h.svc.Invite(c.Request.Context(), middleware.GetOrganizationID(c), req.Email, req.Password)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, profile)
}

func (h *StaffHandler) Remove(c *gin.Context) {
	if !middleware.IsStaffOwner(c) {
		response.Error(c, apperror.Forbidden("only the organization owner can remove staff"))
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.Remove(c.Request.Context(), middleware.GetOrganizationID(c), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

func (h *AuthHandler) Health(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}
