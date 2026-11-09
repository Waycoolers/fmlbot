package middlewares

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

func secretMatches(got, want string) bool {
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

// InternalAuthMiddlewareWithPublicPaths требует X-Internal-Secret для ВСЕХ путей, КРОМЕ указанных публичных
func InternalAuthMiddlewareWithPublicPaths(secret string, publicPaths ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Пропускаем публичные пути
			for _, path := range publicPaths {
				parts := strings.SplitN(path, " ", 2)
				if len(parts) != 2 {
					continue
				}
				method, url := parts[0], parts[1]

				if r.URL.Path == url && r.Method == method {
					next.ServeHTTP(w, r)
					return
				}
			}

			if !secretMatches(r.Header.Get("X-Internal-Secret"), secret) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// InternalAuthMiddlewareWithPrivatePaths требует X-Internal-Secret ТОЛЬКО для указанных приватных путей
func InternalAuthMiddlewareWithPrivatePaths(secret string, privatePaths ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем только приватные пути
			for _, path := range privatePaths {
				parts := strings.SplitN(path, " ", 2)
				if len(parts) != 2 {
					continue
				}
				method, url := parts[0], parts[1]

				if r.URL.Path == url && r.Method == method {
					if !secretMatches(r.Header.Get("X-Internal-Secret"), secret) {
						http.Error(w, "Forbidden", http.StatusForbidden)
						return
					}
					next.ServeHTTP(w, r)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// InternalSecretBypass направляет запрос сразу в bypass, если он идёт на указанный префикс
// и содержит верный X-Internal-Secret. Остальные запросы проходят обычную цепочку next.
func InternalSecretBypass(secret, method, prefix string, bypass http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == method && strings.HasPrefix(r.URL.Path, prefix) &&
				secretMatches(r.Header.Get("X-Internal-Secret"), secret) {
				bypass.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// LimitBody ограничивает размер тела запроса.
func LimitBody(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
