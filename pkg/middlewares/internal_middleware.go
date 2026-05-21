package middlewares

import (
	"net/http"
	"strings"
)

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

			token := r.Header.Get("X-Internal-Secret")
			if token != secret {
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
					token := r.Header.Get("X-Internal-Secret")
					if token != secret {
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
