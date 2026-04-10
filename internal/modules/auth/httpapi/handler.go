package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"auth-service/internal/core/database"
	"auth-service/internal/core/httpx"
	"auth-service/internal/core/jwtx"
	"auth-service/internal/modules/auth/model"
	"auth-service/internal/modules/auth/repository"
	"auth-service/internal/modules/auth/service"
	"github.com/gorilla/mux"
)

type Handler struct {
	dbManager *database.Manager
	jwt       *jwtx.Service
}

func NewHandler(dbManager *database.Manager, jwtService *jwtx.Service) *Handler {
	return &Handler{dbManager: dbManager, jwt: jwtService}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	httpx.SendSuccess(w, http.StatusOK, "auth service healthy", nil)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	svc, domain, err := h.serviceFromRequest(r)
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.SendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := svc.Register(req)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "already") {
			status = http.StatusConflict
		}
		httpx.SendError(w, status, err.Error())
		return
	}

	httpx.SendSuccess(w, http.StatusCreated, "user registered successfully", map[string]interface{}{
		"domain": domain,
		"user":   user,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	svc, domain, err := h.serviceFromRequest(r)
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.SendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := svc.Login(domain, req)
	if err != nil {
		httpx.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	httpx.SendSuccess(w, http.StatusOK, "login successful", resp)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	svc, _, err := h.serviceFromRequest(r)
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req model.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.SendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	uid := httpx.UserIDFromContext(r.Context())
	if uid == 0 {
		httpx.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := svc.ChangePassword(uid, req.OldPassword, req.NewPassword); err != nil {
		httpx.SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	httpx.SendSuccess(w, http.StatusOK, "password changed successfully", nil)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	svc, _, err := h.serviceFromRequest(r)
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	uid := httpx.UserIDFromContext(r.Context())
	if uid == 0 {
		httpx.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	data, err := svc.Me(uid)
	if err != nil {
		httpx.SendError(w, http.StatusNotFound, err.Error())
		return
	}
	
	httpx.SendSuccess(w, http.StatusOK, "user retrieved successfully", data)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	httpx.SendSuccess(w, http.StatusOK, "logout successful", nil)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	svc, _, err := h.serviceFromRequest(r)
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	users, err := svc.ListUsers()
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}
	httpx.SendSuccess(w, http.StatusOK, "users retrieved successfully", users)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	svc, _, err := h.serviceFromRequest(r)
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	uid, err := idParam(r, "id")
	if err != nil {
		httpx.SendError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	data, err := svc.GetUser(uid)
	if err != nil {
		httpx.SendError(w, http.StatusNotFound, err.Error())
		return
	}
	httpx.SendSuccess(w, http.StatusOK, "user retrieved successfully", data)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	svc, _, err := h.serviceFromRequest(r)
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	uid, err := idParam(r, "id")
	if err != nil {
		httpx.SendError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.SendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := svc.UpdateUser(uid, req); err != nil {
		httpx.SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.SendSuccess(w, http.StatusOK, "user updated successfully", nil)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	svc, _, err := h.serviceFromRequest(r)
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	uid, err := idParam(r, "id")
	if err != nil {
		httpx.SendError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := svc.DeleteUser(uid); err != nil {
		httpx.SendError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}
	httpx.SendSuccess(w, http.StatusOK, "user deleted successfully", nil)
}

func (h *Handler) AssignRole(w http.ResponseWriter, r *http.Request) {
	svc, _, err := h.serviceFromRequest(r)
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	uid, err := idParam(r, "id")
	if err != nil {
		httpx.SendError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req model.AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.SendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RoleID == 0 {
		httpx.SendError(w, http.StatusBadRequest, "role_id is required")
		return
	}

	if err := svc.AssignRole(uid, req.RoleID); err != nil {
		httpx.SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.SendSuccess(w, http.StatusOK, "role assigned successfully", nil)
}

func (h *Handler) RemoveRole(w http.ResponseWriter, r *http.Request) {
	svc, _, err := h.serviceFromRequest(r)
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	uid, err := idParam(r, "id")
	if err != nil {
		httpx.SendError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	roleID, err := idParam(r, "roleId")
	if err != nil {
		httpx.SendError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	if err := svc.RemoveRole(uid, roleID); err != nil {
		httpx.SendError(w, http.StatusInternalServerError, "failed to remove role")
		return
	}
	httpx.SendSuccess(w, http.StatusOK, "role removed successfully", nil)
}

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	svc, _, err := h.serviceFromRequest(r)
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	roles, err := svc.ListRoles()
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, "failed to fetch roles")
		return
	}
	httpx.SendSuccess(w, http.StatusOK, "roles retrieved successfully", roles)
}

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	svc, _, err := h.serviceFromRequest(r)
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.SendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	role, err := svc.CreateRole(body.Name, body.Description)
	if err != nil {
		httpx.SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.SendSuccess(w, http.StatusCreated, "role created successfully", role)
}

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	svc, _, err := h.serviceFromRequest(r)
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	permissions, err := svc.ListPermissions()
	if err != nil {
		httpx.SendError(w, http.StatusInternalServerError, "failed to fetch permissions")
		return
	}
	httpx.SendSuccess(w, http.StatusOK, "permissions retrieved successfully", permissions)
}

func (h *Handler) serviceFromRequest(r *http.Request) (*service.Service, string, error) {
	domain := httpx.DomainFromContext(r.Context())
	if domain == "" {
		domain = "default"
	}

	if err := h.dbManager.EnsureDomainReady(domain); err != nil {
		return nil, "", err
	}

	db, resolvedDomain, err := h.dbManager.DBForDomain(domain)
	if err != nil {
		return nil, "", err
	}

	repo := repository.New(db)
	return service.New(repo, h.jwt), resolvedDomain, nil
}

func idParam(r *http.Request, key string) (uint, error) {
	v := mux.Vars(r)[key]
	id, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
