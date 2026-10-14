package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Claims struct {
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	Role      Role      `json:"role"`
	ExpiresAt time.Time `json:"exp"`
}

type TokenService struct {
	secretKey []byte
}

func NewTokenService(secret string) *TokenService {
	if secret == "" {
		secret = "pulsegrid-enterprise-default-secret-key-32b"
	}
	return &TokenService{secretKey: []byte(secret)}
}

func (s *TokenService) GenerateToken(tenantID, userID string, role Role, duration time.Duration) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, _ := json.Marshal(header)
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	claims := Claims{
		TenantID:  tenantID,
		UserID:    userID,
		Role:      role,
		ExpiresAt: time.Now().Add(duration),
	}
	claimsJSON, _ := json.Marshal(claims)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	payload := fmt.Sprintf("%s.%s", headerEncoded, claimsEncoded)
	sig := s.sign(payload)
	return fmt.Sprintf("%s.%s", payload, sig), nil
}

func (s *TokenService) ValidateToken(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	payload := fmt.Sprintf("%s.%s", parts[0], parts[1])
	expectedSig := s.sign(payload)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, errors.New("token signature verification failed")
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid claims encoding")
	}

	var claims Claims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, errors.New("failed to parse claims")
	}

	if time.Now().After(claims.ExpiresAt) {
		return nil, errors.New("token has expired")
	}

	return &claims, nil
}

func (s *TokenService) sign(data string) string {
	h := hmac.New(sha256.New, s.secretKey)
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
