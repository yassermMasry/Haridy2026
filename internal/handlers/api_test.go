package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"haridy2026/configs"
	"haridy2026/internal/models"
	"haridy2026/internal/routes"
	"haridy2026/internal/testutil"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func TestAPIItemsAreTenantIsolated(t *testing.T) {
	testutil.ChdirRepoRoot(t)
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
	testutil.ChdirRepoRoot(t)
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

func TestWebLoginPersistsSessionWithRotatedSecrets(t *testing.T) {
	testutil.ChdirRepoRoot(t)
	fx := testutil.NewFixture(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if err := fx.DB.Model(&fx.User).Update("password_hash", string(hash)).Error; err != nil {
		t.Fatalf("update password: %v", err)
	}
	router := routes.Setup(fx.DB, configs.Config{
		AppEnv:        "development",
		AppSecret:     "change-this-secret-in-production",
		AppSecrets:    []string{"change-this-secret-in-production", "previous-secret-during-rotation"},
		JWTSecret:     "jwt-secret",
		JWTSecrets:    []string{"jwt-secret"},
		SessionName:   "haridy_session",
		SessionMaxAge: 28800,
		SessionSecure: false,
	})

	form := url.Values{"username": {"admin"}, "password": {"admin123"}}
	login := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(login, req)

	if login.Code != http.StatusFound {
		t.Fatalf("expected login redirect, got %d body=%s", login.Code, login.Body.String())
	}
	if location := login.Header().Get("Location"); location != "/dashboard" {
		t.Fatalf("expected redirect to /dashboard, got %q", location)
	}
	cookies := login.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected login to set session cookie")
	}
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "haridy_session" {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("missing haridy_session cookie; got %v", cookies)
	}
	if sessionCookie.Path != "/" || !sessionCookie.HttpOnly || sessionCookie.Secure {
		t.Fatalf("unexpected local session cookie options: path=%q httponly=%t secure=%t", sessionCookie.Path, sessionCookie.HttpOnly, sessionCookie.Secure)
	}

	dashboard := httptest.NewRecorder()
	dashboardReq := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	dashboardReq.AddCookie(sessionCookie)
	router.ServeHTTP(dashboard, dashboardReq)
	if dashboard.Code != http.StatusOK {
		t.Fatalf("expected dashboard 200 with saved session, got %d location=%q body=%s", dashboard.Code, dashboard.Header().Get("Location"), dashboard.Body.String())
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
