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
	Config  *config.Config
	Log     *slog.Logger
	Tokens  *auth.TokenService
	Auth    *service.AuthService
	Settings *service.SettingsService
	Rooms   *service.RoomService
	Tenants *service.TenantService
	Payments *service.PaymentService
	Expenses *service.ExpenseService
	Kitchen *service.KitchenService
	Staff   *service.StaffService
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
	roomH := handler.NewRoomHandler(deps.Rooms)
	tenantH := handler.NewTenantHandler(deps.Tenants)
	paymentH := handler.NewPaymentHandler(deps.Payments)
	expenseH := handler.NewExpenseHandler(deps.Expenses)
	kitchenH := handler.NewKitchenHandler(deps.Kitchen)
	staffH := handler.NewStaffHandler(deps.Staff)

	r.GET("/health", h.Health)

	v1 := r.Group("/api/v1")
	{
		// Public auth
		v1.POST("/auth/staff/login", h.StaffLogin)
		v1.POST("/auth/tenant/login", h.TenantLogin)

		// Staff routes
		staff := v1.Group("", middleware.StaffAuth(deps.Tokens))
		{
			staff.POST("/auth/staff/logout", h.StaffLogout)
			staff.GET("/auth/staff/me", h.StaffMe)

			staff.GET("/settings", settingsH.Get)
			staff.PATCH("/settings", settingsH.Update)

			staff.GET("/rooms", roomH.List)
			staff.POST("/rooms", roomH.Create)
			staff.DELETE("/rooms/:id", roomH.Delete)

			staff.GET("/tenants", tenantH.List)
			staff.POST("/tenants", tenantH.Create)
			staff.POST("/tenants/:id/move-out", tenantH.MoveOut)

			staff.GET("/payments", paymentH.List)
			staff.POST("/payments", paymentH.Create)
			staff.DELETE("/payments/:id", paymentH.Delete)

			staff.GET("/expenses", expenseH.List)
			staff.POST("/expenses", expenseH.Create)
			staff.DELETE("/expenses/:id", expenseH.Delete)

			staff.GET("/kitchen/items", kitchenH.ListItems)
			staff.POST("/kitchen/items", kitchenH.CreateItem)
			staff.POST("/kitchen/items/:id/stock-in", kitchenH.StockIn)
			staff.POST("/kitchen/items/:id/use", kitchenH.UseStock)
			staff.GET("/kitchen/log", kitchenH.ListLog)

			staff.GET("/staff", staffH.List)
			staff.POST("/staff/invite", staffH.Invite)
			staff.DELETE("/staff/:id", staffH.Remove)
		}

		// Tenant routes
		tenant := v1.Group("", middleware.TenantAuth(deps.Tokens))
		{
			tenant.GET("/auth/tenant/me", h.TenantMe)
			tenant.GET("/tenant/payments", paymentH.TenantPayments)
		}
	}

	return r
}
