package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Waycoolers/fmlbot/pkg/errs"
)

type authStub struct{}

func (authStub) GetAccessToken(context.Context, int64) (string, error) { return "token", nil }

func clientFor(t *testing.T, h http.HandlerFunc) *client {
	t.Helper()
	server := httptest.NewServer(h)
	t.Cleanup(server.Close)
	return &client{baseURL: server.URL, httpClient: server.Client(), authClient: authStub{}}
}

func TestDailyLimitIsNotReportedAsEmptyBucket(t *testing.T) {
	c := clientFor(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": errs.ErrLimitExceeded.Error()})
	})
	_, err := c.ReceiveNextCompliment(context.Background(), 1)
	if !errors.Is(err, errs.ErrLimitExceeded) {
		t.Fatalf("got %v", err)
	}
}

func TestEmptyBucketKeepsMinutes(t *testing.T) {
	c := clientFor(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "bucket is empty", "minutes": 17})
	})
	_, err := c.ReceiveNextCompliment(context.Background(), 1)
	var bucket *errs.ErrBucketEmpty
	if !errors.As(err, &bucket) || bucket.Minutes != 17 {
		t.Fatalf("got %v", err)
	}
}

func TestCreateUserConflict(t *testing.T) {
	c := clientFor(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, errs.ErrUsernameIsAlreadyTaken.Error(), http.StatusConflict)
	})
	_, err := c.CreateUser(context.Background(), 1, "name")
	if !errors.Is(err, errs.ErrUserExists) {
		t.Fatalf("got %v", err)
	}
}

func TestPairWithTakenPartner(t *testing.T) {
	c := clientFor(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": errs.ErrPartnerAlreadyHasPartner.Error()})
	})
	if _, err := c.RequestPartner(context.Background(), 1, 2); !errors.Is(err, errs.ErrPartnerAlreadyHasPartner) {
		t.Fatalf("got %v", err)
	}
}

func TestUsernameIsEscapedInPath(t *testing.T) {
	var path string
	c := clientFor(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNotFound)
	})
	_, _ = c.GetUserByUsername(context.Background(), 1, "../me?x=1")
	if path != "/users/by-username/..%2Fme%3Fx=1" {
		t.Fatalf("got %q", path)
	}
}
