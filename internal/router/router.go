package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/nivas/server/internal/auth"
	"github.com/nivas/server/internal/config"
	"github.com/nivas/server/internal/handler"
	"github.com/nivas/server/internal/middleware"
	"github.com/nivas/server/internal/service"
)

type Deps struct {
	Config    *config.Config
	Log       *slog.Logger
	Tokens    *auth.TokenService
	Auth      *service.AuthService
	Settings  *service.SettingsService
	Properties *service.PropertyService
	Rooms     *service.RoomService
	Tenants   *service.TenantService
	Payments  *service.PaymentService
	Expenses  *service.ExpenseService
	Kitchen   *service.KitchenService
	Staff     *service.StaffService
	Audit     *service.AuditService
	Export    *service.ExportService
	Reminders *service.ReminderService
	Documents *service.DocumentService
	Portal    *service.PortalService
}

func New(deps Deps) *gin.Engine {
	if deps.Config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Recovery(deps.Log))
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(deps.Log))
	r.Use(middleware.CORS(deps.Config.CORS.AllowedOrigins))

	h := handler.NewAuthHandler(deps.Auth)
	settingsH := handler.NewSettingsHandler(deps.Settings)
	propertyH := handler.NewPropertyHandler(deps.Properties)
	roomH := handler.NewRoomHandler(deps.Rooms)
	tenantH := handler.NewTenantHandler(deps.Tenants)
	paymentH := handler.NewPaymentHandler(deps.Payments)
	expenseH := handler.NewExpenseHandler(deps.Expenses)
	kitchenH := handler.NewKitchenHandler(deps.Kitchen)
	staffH := handler.NewStaffHandler(deps.Staff)
	auditH := handler.NewAuditHandler(deps.Audit)
	exportH := handler.NewExportHandler(deps.Export)
	reminderH := handler.NewReminderHandler(deps.Reminders)
	documentH := handler.NewDocumentHandler(deps.Documents)
	portalH := handler.NewPortalHandler(deps.Portal)

	authRL := middleware.NewRateLimiter(deps.Config.RateLimit.AuthLimit, deps.Config.RateLimit.AuthWindow)

	r.GET("/health", h.Health)

	v1 := r.Group("/api/v1")
	{
		authRoutes := v1.Group("", authRL.Middleware())
		{
			authRoutes.POST("/auth/staff/login", h.StaffLogin)
			authRoutes.POST("/auth/staff/refresh", h.StaffRefresh)
			authRoutes.POST("/auth/staff/logout", h.StaffLogout)
			authRoutes.POST("/auth/staff/forgot-password", h.StaffForgotPassword)
			authRoutes.POST("/auth/staff/reset-password", h.StaffResetPassword)
			authRoutes.POST("/auth/tenant/login", h.TenantLogin)
			authRoutes.POST("/auth/tenant/refresh", h.TenantRefresh)
			authRoutes.POST("/auth/tenant/logout", h.TenantLogout)
			authRoutes.POST("/auth/tenant/forgot-password", h.TenantForgotPassword)
			authRoutes.POST("/auth/tenant/reset-password", h.TenantResetPassword)
		}

		staff := v1.Group("", middleware.StaffAuth(deps.Tokens))
		{
			staff.GET("/auth/staff/me", h.StaffMe)

			staff.GET("/settings", settingsH.Get)
			staff.PATCH("/settings", settingsH.Update)

			staff.GET("/properties", propertyH.List)
			staff.POST("/properties", propertyH.Create)
			staff.PATCH("/properties/:id", propertyH.Update)

			staff.GET("/rooms", roomH.List)
			staff.POST("/rooms", roomH.Create)
			staff.DELETE("/rooms/:id", roomH.Delete)

			staff.GET("/tenants", tenantH.List)
			staff.POST("/tenants", tenantH.Create)
			staff.POST("/tenants/:id/move-out", tenantH.MoveOut)

			staff.GET("/payments", paymentH.List)
			staff.POST("/payments", paymentH.Create)
			staff.DELETE("/payments/:id", paymentH.Delete)
			staff.GET("/payments/export", exportH.Payments)

			staff.GET("/expenses", expenseH.List)
			staff.POST("/expenses", expenseH.Create)
			staff.DELETE("/expenses/:id", expenseH.Delete)
			staff.GET("/expenses/export", exportH.Expenses)

			staff.GET("/tenants/export", exportH.Tenants)

			staff.POST("/reminders/run", reminderH.Run)

			staff.GET("/tenants/:id/documents", documentH.ListTenantDocuments)
			staff.POST("/tenants/:id/documents", documentH.UploadTenantDocument)
			staff.GET("/tenant-documents/:id/download", documentH.DownloadTenantDocument)
			staff.DELETE("/tenant-documents/:id", documentH.DeleteTenantDocument)

			staff.GET("/organization-documents", documentH.ListOrganizationDocuments)
			staff.POST("/organization-documents", documentH.UploadOrganizationDocument)
			staff.GET("/organization-documents/:id/download", documentH.DownloadOrganizationDocument)
			staff.DELETE("/organization-documents/:id", documentH.DeleteOrganizationDocument)

			staff.GET("/announcements", portalH.ListAnnouncements)
			staff.POST("/announcements", portalH.CreateAnnouncement)
			staff.PATCH("/announcements/:id", portalH.UpdateAnnouncement)
			staff.DELETE("/announcements/:id", portalH.DeleteAnnouncement)

			staff.GET("/maintenance-requests", portalH.ListMaintenance)
			staff.PATCH("/maintenance-requests/:id", portalH.UpdateMaintenance)

			staff.GET("/visitor-log", portalH.ListVisitors)
			staff.POST("/visitor-log", portalH.CreateVisitorEntry)
			staff.POST("/visitor-log/:id/exit", portalH.RecordVisitorExit)

			staff.GET("/kitchen/items", kitchenH.ListItems)
			staff.POST("/kitchen/items", kitchenH.CreateItem)
			staff.POST("/kitchen/items/:id/stock-in", kitchenH.StockIn)
			staff.POST("/kitchen/items/:id/use", kitchenH.UseStock)
			staff.GET("/kitchen/log", kitchenH.ListLog)

			staff.GET("/audit-log", auditH.List)

			staff.GET("/staff", staffH.List)
			staff.POST("/staff/invite", staffH.Invite)
			staff.DELETE("/staff/:id", staffH.Remove)
		}

		tenant := v1.Group("", middleware.TenantAuth(deps.Tokens))
		{
			tenant.GET("/auth/tenant/me", h.TenantMe)
			tenant.GET("/tenant/payments", paymentH.TenantPayments)
			tenant.GET("/tenant/announcements", portalH.TenantAnnouncements)
			tenant.GET("/tenant/maintenance-requests", portalH.TenantMaintenance)
			tenant.POST("/tenant/maintenance-requests", portalH.CreateMaintenance)
		}
	}

	return r
}
