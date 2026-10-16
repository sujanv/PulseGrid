package pki

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type ChallengeType string

const (
	ChallengeHTTP01 ChallengeType = "http-01"
	ChallengeDNS01  ChallengeType = "dns-01"
)

type CertificateStatus string

const (
	CertStatusPending    CertificateStatus = "PENDING"
	CertStatusValidating CertificateStatus = "VALIDATING"
	CertStatusIssued     CertificateStatus = "ISSUED"
	CertStatusRevoked    CertificateStatus = "REVOKED"
)

type ManagedCertificate struct {
	Domain        string            `json:"domain"`
	Status        CertificateStatus `json:"status"`
	ChallengeType ChallengeType     `json:"challenge_type"`
	IssuedAt      time.Time         `json:"issued_at"`
	ExpiresAt     time.Time         `json:"expires_at"`
	Fingerprint   string            `json:"fingerprint"`
	AutoRenew     bool              `json:"auto_renew"`
}

type ACMEManager struct {
	certs map[string]*ManagedCertificate
}

func NewACMEManager() *ACMEManager {
	return &ACMEManager{
		certs: make(map[string]*ManagedCertificate),
	}
}

func (m *ACMEManager) RequestCertificate(domain string, challenge ChallengeType) (*ManagedCertificate, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	now := time.Now().UTC()
	fpHash := sha256.Sum256([]byte(domain + now.String()))
	fp := hex.EncodeToString(fpHash[:8])

	cert := &ManagedCertificate{
		Domain:        domain,
		Status:        CertStatusIssued,
		ChallengeType: challenge,
		IssuedAt:      now,
		ExpiresAt:     now.AddDate(0, 3, 0), // 90 days validity
		Fingerprint:   fp,
		AutoRenew:     true,
	}

	m.certs[domain] = cert
	return cert, nil
}

func (m *ACMEManager) CheckRenewalNeeded(domain string, thresholdDays int) (bool, *ManagedCertificate) {
	cert, ok := m.certs[domain]
	if !ok {
		return false, nil
	}

	daysRemaining := int(time.Until(cert.ExpiresAt).Hours() / 24)
	return daysRemaining <= thresholdDays, cert
}
