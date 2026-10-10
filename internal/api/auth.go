package api

import (
	"net/http"
	"strings"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
)

type AuthManager struct {
	adminKey  string
	viewerKey string
}

func NewAuthManager(adminKey, viewerKey string) *AuthManager {
	return &AuthManager{
		adminKey:  adminKey,
		viewerKey: viewerKey,
	}
}

func (a *AuthManager) Authenticate(r *http.Request) (Role, bool) {
	// If no keys configured, allow all
	if a.adminKey == "" && a.viewerKey == "" {
		return RoleAdmin, true
	}

	key := r.Header.Get("X-API-Key")
	if key == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			key = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if key == "" {
		return "", false
	}

	if a.adminKey != "" && key == a.adminKey {
		return RoleAdmin, true
	}
	if a.viewerKey != "" && key == a.viewerKey {
		return RoleViewer, true
	}

	return "", false
}
