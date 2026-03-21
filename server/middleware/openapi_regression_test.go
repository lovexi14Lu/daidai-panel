package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"daidai-panel/middleware"
	"daidai-panel/testutil"

	"github.com/gin-gonic/gin"
)

func TestAppTokenDefaultDenyWithoutOpenAPIAccess(t *testing.T) {
	testutil.SetupTestEnv(t)

	token := testutil.MustCreateAppToken(t, "deny-without-scope", "tasks")
	engine := gin.New()

	handlerReached := false
	engine.GET("/private", middleware.JWTAuth(), middleware.RequireRole("viewer"), func(c *gin.Context) {
		handlerReached = true
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	if handlerReached {
		t.Fatal("handler should not run when app token is not explicitly authorized")
	}
}

func TestAppTokenScopeCanPassOperatorRoute(t *testing.T) {
	testutil.SetupTestEnv(t)

	token := testutil.MustCreateAppToken(t, "operator-scope", "tasks")
	engine := gin.New()

	handlerReached := false
	engine.POST("/operator", middleware.JWTAuth(), middleware.OpenAPIAccess("tasks"), middleware.RequireRole("operator"), func(c *gin.Context) {
		handlerReached = true
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/operator", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !handlerReached {
		t.Fatal("handler should run when app token scope is explicitly authorized")
	}
}
