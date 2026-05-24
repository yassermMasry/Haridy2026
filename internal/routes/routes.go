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
	r.Use(gin.Recovery(), gin.Logger(), middleware.Observability(), middleware.SecureHeaders(), middleware.RateLimit(180, 60_000_000_000), middleware.TenantResolver(db))
	store := cookie.NewStore(sessionKeyPairs(cfg.AppSecrets)...)
	store.Options(sessions.Options{Path: "/", MaxAge: sessionMaxAge(cfg), HttpOnly: true, Secure: cfg.SessionSecure, SameSite: http.SameSiteLaxMode})
	r.Use(sessions.Sessions(cfg.SessionName, store))
	r.Use(middleware.CSRFMiddleware(), middleware.ViewData())
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/**/*")

	authService := services.NewAuthService(db)
	itemService := services.NewItemService(db)
	itemCategoryService := services.NewItemCategoryService(db)
	customerService := services.NewCustomerService(db)
	supplierService := services.NewSupplierService(db)
	accountingService := services.NewAccountingService(db)
	erpService := services.NewERPService(db)
	returnService := services.NewReturnService(db)
	commercialService := services.NewCommercialService(db)
	auth := handlers.NewAuthHandler(authService)
	dashboard := handlers.NewDashboardHandler(services.NewDashboardService(db))
	items := handlers.NewItemHandler(itemService)
	itemCategories := handlers.NewItemCategoryHandler(itemCategoryService)
	sales := handlers.NewSalesHandler(services.NewSalesService(db), itemService, customerService)
	treasury := handlers.NewTreasuryHandler(services.NewTreasuryService(db))
	customers := handlers.NewCustomerHandler(customerService)
	suppliers := handlers.NewSupplierHandler(supplierService)
	purchases := handlers.NewPurchaseHandler(services.NewPurchaseService(db), itemService, supplierService)
	reports := handlers.NewReportHandler(db, accountingService)
	erp := handlers.NewERPHandler(erpService, itemService, customerService, supplierService, returnService)
	redisClient := cache.NewRedis(cfg)
	api := handlers.NewAPIHandler(db, authService, cfg.JWTSecret, redisClient)
	commercial := handlers.NewCommercialHandler(commercialService)
	activation := handlers.NewActivationHandler(services.NewActivationService(db))

	r.GET("/", commercial.Landing)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/readyz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ready"}) })
	r.GET("/metrics", func(c *gin.Context) { c.String(http.StatusOK, middleware.MetricsText()) })
	r.GET("/onboarding", commercial.Onboarding)
	r.POST("/onboarding", commercial.CreateCompany)
	r.GET("/login", auth.LoginPage)
	r.POST("/login", auth.Login)
	r.POST("/logout", auth.Logout)

	protected := r.Group("/")
	protected.Use(middleware.AuthRequired())
	protected.GET("/dashboard", dashboard.Index)
	protected.GET("/activation", activation.Show)
	protected.POST("/activation", activation.Activate)
	protected.GET("/price-list", items.PriceList)
	protected.POST("/price-list/items/:id/update-price", items.UpdatePrice)
	protected.GET("/items", items.Index)
	protected.GET("/items/new", items.Create)
	protected.POST("/items", items.Store)
	protected.GET("/items/:id/edit", items.Edit)
	protected.POST("/items/:id", items.Update)
	protected.POST("/items/:id/delete", items.Delete)
	protected.GET("/item-categories", itemCategories.Index)
	protected.POST("/item-categories", itemCategories.Store)
	protected.POST("/item-categories/quick-create", itemCategories.QuickCreate)
	protected.GET("/item-categories/:id/edit", itemCategories.Edit)
	protected.POST("/item-categories/:id", itemCategories.Update)
	protected.POST("/item-categories/:id/delete", itemCategories.Delete)
	protected.GET("/sales", sales.Index)
	protected.GET("/sales/new", sales.Create)
	protected.POST("/sales", sales.Store)
	protected.GET("/sales/:id", sales.Show)
	protected.GET("/customers", customers.Index)
	protected.GET("/customers/new", customers.Create)
	protected.POST("/customers", customers.Store)
	protected.POST("/customers/quick-create", customers.QuickCreate)
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
	protected.POST("/warehouses", middleware.PermissionRequired(db, "admin.manage"), erp.CreateWarehouse)
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
	protected.GET("/financial/trial-balance", commercial.TrialBalance)
	protected.GET("/financial/statements", commercial.FinancialStatements)
	protected.POST("/financial/close-year", commercial.CloseYear)
	protected.POST("/einvoices/:invoice_id", commercial.EInvoice)

	v1 := r.Group("/api/v1")
	v1.POST("/auth/login", api.Login)
	v1.POST("/auth/refresh", api.Refresh)
	v1.Use(middleware.APIAuth(cfg.JWTSecrets...))
	v1.POST("/mobile/devices", api.RegisterDevice)
	v1.GET("/items", api.ListItems)
	v1.GET("/sales", api.ListSales)
	v1.GET("/purchases", api.ListPurchases)
	v1.GET("/customers", api.ListCustomers)
	v1.GET("/suppliers", api.ListSuppliers)
	v1.GET("/treasury", api.Treasury)
	v1.GET("/reports", api.Reports)
	v1.GET("/financial/statements", commercial.FinancialStatements)
	v1.GET("/financial/trial-balance", commercial.TrialBalance)
	v1.POST("/einvoices/:invoice_id", commercial.EInvoice)

	return r
}

func sessionKeyPairs(secrets []string) [][]byte {
	if len(secrets) == 0 {
		secrets = []string{"change-me"}
	}
	keyPairs := make([][]byte, 0, len(secrets)*2)
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		keyPairs = append(keyPairs, []byte(secret), nil)
	}
	if len(keyPairs) == 0 {
		return [][]byte{[]byte("change-me"), nil}
	}
	return keyPairs
}

func sessionMaxAge(cfg configs.Config) int {
	if cfg.SessionMaxAge > 0 {
		return cfg.SessionMaxAge
	}
	return 28800
}
