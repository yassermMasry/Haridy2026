package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

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
	user, err := h.auth.Login(req.Username, req.Password)
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
	c.JSON(http.StatusOK, gin.H{"token": signed})
}

func (h *APIHandler) ListItems(c *gin.Context) {
	var rows []models.Item
	h.db.Limit(100).Order("id desc").Find(&rows)
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func (h *APIHandler) ListSales(c *gin.Context) {
	var rows []models.SalesInvoice
	h.db.Preload("Customer").Limit(100).Order("id desc").Find(&rows)
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func (h *APIHandler) ListPurchases(c *gin.Context) {
	var rows []models.PurchaseInvoice
	h.db.Preload("Supplier").Limit(100).Order("id desc").Find(&rows)
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func (h *APIHandler) ListCustomers(c *gin.Context) {
	var rows []models.Customer
	h.db.Limit(100).Order("id desc").Find(&rows)
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func (h *APIHandler) ListSuppliers(c *gin.Context) {
	var rows []models.Supplier
	h.db.Limit(100).Order("id desc").Find(&rows)
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
