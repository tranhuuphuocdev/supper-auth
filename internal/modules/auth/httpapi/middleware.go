package httpapi

import (
	"net/http"
	"strings"

	"auth-service/internal/core/config"
	"auth-service/internal/core/database"
	"auth-service/internal/core/httpx"
	"auth-service/internal/core/jwtx"
	"auth-service/internal/modules/auth/repository"
	"auth-service/internal/modules/auth/service"
)

func DomainMiddleware(cfg *config.Config) muxMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			domain := strings.TrimSpace(strings.ToLower(r.Header.Get("X-Project-Domain")))
			if domain == "" {
				domain = cfg.DefaultDomain
			}
			ctx := httpx.WithDomain(r.Context(), domain)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AuthMiddleware(jwtService *jwtx.Service) muxMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := bearerToken(r)
			if tokenString == "" {
				httpx.SendError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			claims, err := jwtService.Validate(tokenString)
			if err != nil {
				httpx.SendError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			ctx := httpx.WithUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func PermissionMiddleware(dbManager *database.Manager, jwtService *jwtx.Service, permission string) muxMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := bearerToken(r)
			if tokenString == "" {
				httpx.SendError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			claims, err := jwtService.Validate(tokenString)
			if err != nil {
				httpx.SendError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			domain := httpx.DomainFromContext(r.Context())
			if domain == "" {
				domain = claims.Domain
			}

			if err := dbManager.EnsureDomainReady(domain); err != nil {
				httpx.SendError(w, http.StatusInternalServerError, "failed to prepare domain database")
				return
			}

			db, _, err := dbManager.DBForDomain(domain)
			if err != nil {
				httpx.SendError(w, http.StatusInternalServerError, "failed to resolve database")
				return
			}

			svc := service.New(repository.New(db), jwtService)
			ok, err := svc.HasPermission(claims.UserID, permission)
			if err != nil {
				httpx.SendError(w, http.StatusInternalServerError, "permission check failed")
				return
			}
			if !ok {
				httpx.SendError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			ctx := httpx.WithUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) string {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader == "" {
		return ""
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}
