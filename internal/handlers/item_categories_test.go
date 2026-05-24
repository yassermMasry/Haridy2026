package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"haridy2026/internal/middleware"
	"haridy2026/internal/services"
	"haridy2026/internal/testutil"

	"github.com/gin-gonic/gin"
)

func TestQuickCreateItemCategoryReturnsJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fx := testutil.NewFixture(t)
	handler := NewItemCategoryHandler(services.NewItemCategoryService(fx.DB))

	form := url.Values{"name": {"منظفات"}}
	req := httptest.NewRequest(http.MethodPost, "/item-categories/quick-create", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(middleware.TenantContextKey, fx.Tenant.ID)

	handler.QuickCreate(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ID == 0 || body.Name != "منظفات" {
		t.Fatalf("unexpected response: %+v", body)
	}
}
