package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"haridy2026/configs"
	"haridy2026/internal/models"
	"haridy2026/internal/routes"
	"haridy2026/internal/testutil"

	"github.com/golang-jwt/jwt/v5"
)

func TestAPIItemsAreTenantIsolated(t *testing.T) {
	fx := testutil.NewFixture(t)
	other := models.Tenant{Name: "Other Co", Slug: "other", Subdomain: "other", Status: "trial"}
	if err := fx.DB.Create(&other).Error; err != nil {
		t.Fatalf("create other tenant: %v", err)
	}
	if err := fx.DB.Create(&models.Item{TenantID: &other.ID, Name: "Other Item", Code: "OTHER-001", SalePrice: 10, Quantity: 9}).Error; err != nil {
		t.Fatalf("create other item: %v", err)
	}

	router := routes.Setup(fx.DB, configs.Config{
		AppEnv:      "test",
		AppSecret:   "test-secret",
		AppSecrets:  []string{"test-secret"},
		JWTSecret:   "jwt-secret",
		JWTSecrets:  []string{"jwt-secret"},
		SessionName: "test-session",
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	req.Header.Set("X-Tenant", "demo")
	req.Header.Set("Authorization", "Bearer "+testJWT(t, fx.User.ID, "jwt-secret"))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", res.Code, res.Body.String())
	}
	var payload struct {
		Data []models.Item `json:"data"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	for _, item := range payload.Data {
		if item.Code == "OTHER-001" {
			t.Fatal("tenant isolated API leaked item from another tenant")
		}
	}
}

func TestSecurityHeadersAndMetrics(t *testing.T) {
	fx := testutil.NewFixture(t)
	router := routes.Setup(fx.DB, configs.Config{AppEnv: "production", AppSecret: "s", AppSecrets: []string{"s"}, JWTSecret: "j", JWTSecrets: []string{"j"}, SessionName: "s"})
	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("expected healthz 200, got %d", res.Code)
	}
	if res.Header().Get("Content-Security-Policy") == "" {
		t.Fatal("missing content security policy")
	}
	if res.Header().Get("Strict-Transport-Security") == "" {
		t.Fatal("missing hsts")
	}
}

func testJWT(t *testing.T, userID uint, secret string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": userID, "exp": time.Now().Add(time.Hour).Unix()})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign jwt: %v", err)
	}
	return signed
}
