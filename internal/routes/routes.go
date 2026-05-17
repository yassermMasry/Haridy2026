package routes

import (
	"net/http"

	"haridy2026/configs"
	"haridy2026/internal/cache"
	"haridy2026/internal/handlers"
	"haridy2026/internal/middleware"
	"haridy2026/internal/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, cfg configs.Config) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), middleware.SecureHeaders(), middleware.RateLimit(180, 60_000_000_000))
	store := cookie.NewStore([]byte(cfg.AppSecret))
	store.Options(sessions.Options{Path: "/", MaxAge: 28800, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	r.Use(sessions.Sessions(cfg.SessionName, store))
	r.Use(middleware.CSRFMiddleware(), middleware.ViewData())
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/**/*")

	authService := services.NewAuthService(db)
	itemService := services.NewItemService(db)
	customerService := services.NewCustomerService(db)
	supplierService := services.NewSupplierService(db)
	accountingService := services.NewAccountingService(db)
	erpService := services.NewERPService(db)
	returnService := services.NewReturnService(db)
	auth := handlers.NewAuthHandler(authService)
	dashboard := handlers.NewDashboardHandler(services.NewDashboardService(db))
	items := handlers.NewItemHandler(itemService)
	sales := handlers.NewSalesHandler(services.NewSalesService(db), itemService, customerService)
	treasury := handlers.NewTreasuryHandler(services.NewTreasuryService(db))
	customers := handlers.NewCustomerHandler(customerService)
	suppliers := handlers.NewSupplierHandler(supplierService)
	purchases := handlers.NewPurchaseHandler(services.NewPurchaseService(db), itemService, supplierService)
	reports := handlers.NewReportHandler(db, accountingService)
	erp := handlers.NewERPHandler(erpService, itemService, customerService, supplierService, returnService)
	redisClient := cache.NewRedis(cfg)
	api := handlers.NewAPIHandler(db, authService, cfg.JWTSecret, redisClient)

	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/dashboard") })
	r.GET("/login", auth.LoginPage)
	r.POST("/login", auth.Login)
	r.POST("/logout", auth.Logout)

	protected := r.Group("/")
	protected.Use(middleware.AuthRequired())
	protected.GET("/dashboard", dashboard.Index)
	protected.GET("/items", items.Index)
	protected.GET("/items/new", items.Create)
	protected.POST("/items", items.Store)
	protected.GET("/items/:id/edit", items.Edit)
	protected.POST("/items/:id", items.Update)
	protected.POST("/items/:id/delete", items.Delete)
	protected.GET("/sales", sales.Index)
	protected.GET("/sales/new", sales.Create)
	protected.POST("/sales", sales.Store)
	protected.GET("/sales/:id", sales.Show)
	protected.GET("/customers", customers.Index)
	protected.GET("/customers/new", customers.Create)
	protected.POST("/customers", customers.Store)
	protected.GET("/customers/:id", customers.Show)
	protected.GET("/customers/:id/edit", customers.Edit)
	protected.POST("/customers/:id", customers.Update)
	protected.POST("/customers/:id/receive", customers.Receive)
	protected.GET("/suppliers", suppliers.Index)
	protected.GET("/suppliers/new", suppliers.Create)
	protected.POST("/suppliers", suppliers.Store)
	protected.GET("/suppliers/:id", suppliers.Show)
	protected.GET("/suppliers/:id/edit", suppliers.Edit)
	protected.POST("/suppliers/:id", suppliers.Update)
	protected.POST("/suppliers/:id/pay", suppliers.Pay)
	protected.GET("/purchases", purchases.Index)
	protected.GET("/purchases/new", purchases.Create)
	protected.POST("/purchases", purchases.Store)
	protected.GET("/purchases/:id", purchases.Show)
	protected.GET("/treasury", treasury.Index)
	protected.POST("/treasury", treasury.Store)
	protected.GET("/reports", reports.Index)
	protected.GET("/reports/export/:name", reports.Export)
	protected.GET("/warehouses", middleware.PermissionRequired(db, "admin.manage"), erp.Warehouses)
	protected.POST("/warehouses/transfer", middleware.PermissionRequired(db, "edit.manage"), erp.Transfer)
	protected.POST("/branch/current", erp.SetBranch)
	protected.GET("/vouchers/receipt/new", erp.NewReceipt)
	protected.POST("/vouchers/receipt", erp.CreateReceipt)
	protected.GET("/vouchers/receipt/:number", erp.PrintReceipt)
	protected.GET("/vouchers/payment/new", erp.NewPayment)
	protected.POST("/vouchers/payment", erp.CreatePayment)
	protected.GET("/vouchers/payment/:number", erp.PrintPayment)
	protected.POST("/returns/sales", erp.SalesReturn)
	protected.POST("/returns/purchases", erp.PurchaseReturn)

	v1 := r.Group("/api/v1")
	v1.POST("/auth/login", api.Login)
	v1.Use(middleware.APIAuth(cfg.JWTSecret))
	v1.GET("/items", api.ListItems)
	v1.GET("/sales", api.ListSales)
	v1.GET("/purchases", api.ListPurchases)
	v1.GET("/customers", api.ListCustomers)
	v1.GET("/suppliers", api.ListSuppliers)
	v1.GET("/treasury", api.Treasury)
	v1.GET("/reports", api.Reports)

	return r
}
