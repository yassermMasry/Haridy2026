package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"haridy2026/internal/middleware"
	"haridy2026/internal/models"
	"haridy2026/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type APIHandler struct {
	db     *gorm.DB
	auth   *services.AuthService
	secret string
	redis  *redis.Client
}

func NewAPIHandler(db *gorm.DB, auth *services.AuthService, secret string, redisClient *redis.Client) *APIHandler {
	return &APIHandler{db: db, auth: auth, secret: secret, redis: redisClient}
}

func (h *APIHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	user, err := h.auth.LoginWithContext(req.Username, req.Password, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(8 * time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(h.secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token signing failed"})
		return
	}
	refresh := randomTokenText()
	hash := tokenHash(refresh)
	tenantID := middleware.CurrentTenantID(c)
	var tenantPtr *uint
	if tenantID > 0 {
		tenantPtr = &tenantID
	}
	_ = h.db.Create(&models.RefreshToken{TenantID: tenantPtr, UserID: user.ID, TokenHash: hash, ExpiresAt: time.Now().Add(30 * 24 * time.Hour)}).Error
	c.JSON(http.StatusOK, gin.H{"token": signed, "refresh_token": refresh})
}

func (h *APIHandler) ListItems(c *gin.Context) {
	var rows []models.Item
	h.tenantQuery(c, h.db).Limit(100).Order("id desc").Find(&rows)
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func (h *APIHandler) ListSales(c *gin.Context) {
	var rows []models.SalesInvoice
	h.tenantQuery(c, h.db).Preload("Customer").Limit(100).Order("id desc").Find(&rows)
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func (h *APIHandler) ListPurchases(c *gin.Context) {
	var rows []models.PurchaseInvoice
	h.tenantQuery(c, h.db).Preload("Supplier").Limit(100).Order("id desc").Find(&rows)
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func (h *APIHandler) ListCustomers(c *gin.Context) {
	var rows []models.Customer
	h.tenantQuery(c, h.db).Limit(100).Order("id desc").Find(&rows)
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func (h *APIHandler) ListSuppliers(c *gin.Context) {
	var rows []models.Supplier
	h.tenantQuery(c, h.db).Limit(100).Order("id desc").Find(&rows)
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func (h *APIHandler) Treasury(c *gin.Context) {
	var rows []models.TreasuryTransaction
	h.db.Limit(100).Order("id desc").Find(&rows)
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func (h *APIHandler) Reports(c *gin.Context) {
	ctx := context.Background()
	if h.redis != nil {
		if cached, err := h.redis.Get(ctx, "api:v1:reports:summary").Result(); err == nil {
			var payload gin.H
			if json.Unmarshal([]byte(cached), &payload) == nil {
				c.JSON(http.StatusOK, gin.H{"data": payload, "cached": true})
				return
			}
		}
	}
	var salesTotal, purchaseTotal float64
	h.db.Model(&models.SalesInvoice{}).Select("COALESCE(SUM(total),0)").Scan(&salesTotal)
	h.db.Model(&models.PurchaseInvoice{}).Select("COALESCE(SUM(total),0)").Scan(&purchaseTotal)
	payload := gin.H{"sales_total": salesTotal, "purchase_total": purchaseTotal, "gross_profit": salesTotal - purchaseTotal}
	if h.redis != nil {
		if raw, err := json.Marshal(payload); err == nil {
			_ = h.redis.Set(ctx, "api:v1:reports:summary", raw, 5*time.Minute).Err()
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": payload, "cached": false})
}

func (h *APIHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	var stored models.RefreshToken
	if err := h.db.Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", tokenHash(req.RefreshToken), time.Now()).First(&stored).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": stored.UserID, "exp": time.Now().Add(8 * time.Hour).Unix()})
	signed, err := token.SignedString([]byte(h.secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token signing failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": signed})
}

func (h *APIHandler) RegisterDevice(c *gin.Context) {
	var req struct {
		Platform  string `json:"platform"`
		PushToken string `json:"push_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	userID, _ := c.Get("api_user_id")
	uid, _ := userID.(uint)
	tenantID := middleware.CurrentTenantID(c)
	device := models.MobileDevice{TenantID: tenantID, UserID: uid, Platform: req.Platform, PushToken: req.PushToken}
	if err := h.db.Create(&device).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": device})
}

func (h *APIHandler) tenantQuery(c *gin.Context, db *gorm.DB) *gorm.DB {
	tenantID := middleware.CurrentTenantID(c)
	if tenantID == 0 {
		return db
	}
	return db.Where("tenant_id = ? OR tenant_id IS NULL", tenantID)
}

func randomTokenText() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func tokenHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
