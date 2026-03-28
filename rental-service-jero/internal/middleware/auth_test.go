package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testSecret = "test-secret-key"

func generateToken(secret string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(testSecret))
	r.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		c.JSON(http.StatusOK, gin.H{"user_id": userID, "role": role})
	})
	return r
}

func setupAdminRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(testSecret))
	r.Use(RequireAdmin())
	r.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	router := setupRouter()
	userID := uuid.New()

	token := generateToken(testSecret, jwt.MapClaims{
		"sub":  userID.String(),
		"type": "access",
		"exp":  time.Now().Add(30 * time.Minute).Unix(),
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	router := setupRouter()

	token := generateToken(testSecret, jwt.MapClaims{
		"sub":  uuid.New().String(),
		"type": "access",
		"exp":  time.Now().Add(-1 * time.Hour).Unix(),
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	router := setupRouter()

	token := generateToken("wrong-secret", jwt.MapClaims{
		"sub":  uuid.New().String(),
		"type": "access",
		"exp":  time.Now().Add(30 * time.Minute).Unix(),
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_WrongTokenType(t *testing.T) {
	router := setupRouter()

	token := generateToken(testSecret, jwt.MapClaims{
		"sub":  uuid.New().String(),
		"type": "refresh",
		"exp":  time.Now().Add(30 * time.Minute).Unix(),
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_MissingSub(t *testing.T) {
	router := setupRouter()

	token := generateToken(testSecret, jwt.MapClaims{
		"type": "access",
		"exp":  time.Now().Add(30 * time.Minute).Unix(),
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ExtractsRole(t *testing.T) {
	router := setupRouter()

	token := generateToken(testSecret, jwt.MapClaims{
		"sub":  uuid.New().String(),
		"type": "access",
		"role": "admin",
		"exp":  time.Now().Add(30 * time.Minute).Unix(),
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"role":"admin"`) {
		t.Errorf("expected role admin in response, got %s", w.Body.String())
	}
}

func TestAuthMiddleware_DefaultsRoleToUser(t *testing.T) {
	router := setupRouter()

	token := generateToken(testSecret, jwt.MapClaims{
		"sub":  uuid.New().String(),
		"type": "access",
		"exp":  time.Now().Add(30 * time.Minute).Unix(),
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"role":"user"`) {
		t.Errorf("expected role user in response, got %s", w.Body.String())
	}
}

func TestRequireAdmin_AllowsAdmin(t *testing.T) {
	router := setupAdminRouter()

	token := generateToken(testSecret, jwt.MapClaims{
		"sub":  uuid.New().String(),
		"type": "access",
		"role": "admin",
		"exp":  time.Now().Add(30 * time.Minute).Unix(),
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireAdmin_DeniesUser(t *testing.T) {
	router := setupAdminRouter()

	token := generateToken(testSecret, jwt.MapClaims{
		"sub":  uuid.New().String(),
		"type": "access",
		"role": "user",
		"exp":  time.Now().Add(30 * time.Minute).Unix(),
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireAdmin_DeniesNoRole(t *testing.T) {
	router := setupAdminRouter()

	token := generateToken(testSecret, jwt.MapClaims{
		"sub":  uuid.New().String(),
		"type": "access",
		"exp":  time.Now().Add(30 * time.Minute).Unix(),
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}
