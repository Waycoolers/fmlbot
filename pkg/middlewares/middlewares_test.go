package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var testSecret = []byte("test-secret")

func ok() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
}

func serve(h http.Handler, r *http.Request) int {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Code
}

func signed(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestJWTRequiresExpiration(t *testing.T) {
	h := Middleware(testSecret)(ok())

	r := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	r.Header.Set("Authorization", "Bearer "+signed(t, jwt.MapClaims{"user_id": 1, "exp": time.Now().Add(time.Hour).Unix()}))
	if code := serve(h, r); code != http.StatusOK {
		t.Fatalf("valid token: got %d", code)
	}

	r = httptest.NewRequest(http.MethodGet, "/users/me", nil)
	r.Header.Set("Authorization", "Bearer "+signed(t, jwt.MapClaims{"user_id": 1}))
	if code := serve(h, r); code != http.StatusUnauthorized {
		t.Fatalf("token without exp: got %d", code)
	}
}

func TestJWTPublicPathIsExact(t *testing.T) {
	h := Middleware(testSecret, "POST /auth/verify")(ok())
	if code := serve(h, httptest.NewRequest(http.MethodPost, "/auth/verify", nil)); code != http.StatusOK {
		t.Fatalf("public path: got %d", code)
	}
	if code := serve(h, httptest.NewRequest(http.MethodPost, "/auth/verify/other", nil)); code != http.StatusUnauthorized {
		t.Fatalf("other path: got %d", code)
	}
}

func TestInternalSecretBypass(t *testing.T) {
	denied := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusUnauthorized) })
	h := InternalSecretBypass("secret", http.MethodGet, "/users/by-username/", ok())(denied)

	r := httptest.NewRequest(http.MethodGet, "/users/by-username/alice", nil)
	if code := serve(h, r); code != http.StatusUnauthorized {
		t.Fatalf("without secret: got %d", code)
	}
	r = httptest.NewRequest(http.MethodGet, "/users/by-username/alice", nil)
	r.Header.Set("X-Internal-Secret", "wrong")
	if code := serve(h, r); code != http.StatusUnauthorized {
		t.Fatalf("wrong secret: got %d", code)
	}
	r = httptest.NewRequest(http.MethodGet, "/users/by-username/alice", nil)
	r.Header.Set("X-Internal-Secret", "secret")
	if code := serve(h, r); code != http.StatusOK {
		t.Fatalf("right secret: got %d", code)
	}
	r = httptest.NewRequest(http.MethodGet, "/users/me", nil)
	r.Header.Set("X-Internal-Secret", "secret")
	if code := serve(h, r); code != http.StatusUnauthorized {
		t.Fatalf("secret must not open other paths: got %d", code)
	}
}
