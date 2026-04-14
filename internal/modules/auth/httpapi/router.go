package httpapi

import (
	"net/http"
	"time"

	"auth-service/internal/core/config"
	"auth-service/internal/core/database"
	"auth-service/internal/core/jwtx"

	"github.com/gorilla/mux"
)

type muxMiddleware = mux.MiddlewareFunc

func RegisterRoutes(router *mux.Router, dbManager *database.Manager, cfg *config.Config) {
	jwtService := jwtx.New(cfg.JWTSecret, 24*time.Hour)
	handler := NewHandler(dbManager, jwtService)

	router.HandleFunc("/health", handler.Health).Methods("GET")

	router.HandleFunc("/api/v1/auth/register", handler.Register).Methods("POST")
	router.HandleFunc("/api/v1/auth/login", handler.Login).Methods("POST")

	router.Handle("/api/v1/auth/change-password", apply(handler.ChangePassword, AuthMiddleware(jwtService))).Methods("POST")
	router.Handle("/api/v1/auth/me", apply(handler.Me, AuthMiddleware(jwtService))).Methods("GET")
	router.Handle("/api/v1/auth/logout", apply(handler.Logout, AuthMiddleware(jwtService))).Methods("POST")

	router.Handle("/api/v1/admin/users", apply(handler.ListUsers, AuthMiddleware(jwtService), PermissionMiddleware(dbManager, jwtService, "user.read"))).Methods("GET")
	router.Handle("/api/v1/admin/users/{id}", apply(handler.GetUser, AuthMiddleware(jwtService), PermissionMiddleware(dbManager, jwtService, "user.read"))).Methods("GET")
	router.Handle("/api/v1/admin/users/{id}", apply(handler.UpdateUser, AuthMiddleware(jwtService), PermissionMiddleware(dbManager, jwtService, "user.update"))).Methods("PUT")
	router.Handle("/api/v1/admin/users/{id}", apply(handler.DeleteUser, AuthMiddleware(jwtService), PermissionMiddleware(dbManager, jwtService, "user.delete"))).Methods("DELETE")

	router.Handle("/api/v1/admin/users/{id}/roles", apply(handler.AssignRole, AuthMiddleware(jwtService), PermissionMiddleware(dbManager, jwtService, "role.manage"))).Methods("POST")
	router.Handle("/api/v1/admin/users/{id}/roles/{roleId}", apply(handler.RemoveRole, AuthMiddleware(jwtService), PermissionMiddleware(dbManager, jwtService, "role.manage"))).Methods("DELETE")

	router.Handle("/api/v1/admin/roles", apply(handler.ListRoles, AuthMiddleware(jwtService), PermissionMiddleware(dbManager, jwtService, "role.manage"))).Methods("GET")
	router.Handle("/api/v1/admin/roles", apply(handler.CreateRole, AuthMiddleware(jwtService), PermissionMiddleware(dbManager, jwtService, "role.manage"))).Methods("POST")

	router.Handle("/api/v1/admin/permissions", apply(handler.ListPermissions, AuthMiddleware(jwtService), PermissionMiddleware(dbManager, jwtService, "permission.manage"))).Methods("GET")
}

func apply(handler http.HandlerFunc, middlewares ...muxMiddleware) http.Handler {
	var h http.Handler = handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
