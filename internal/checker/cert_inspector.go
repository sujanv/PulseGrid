package checker

import (
	"crypto/x509"
	"fmt"
	"strings"
	"time"
)

type ExpiryTier string

const (
	ExpiryTierOK       ExpiryTier = "OK"
	ExpiryTierWarning  ExpiryTier = "EXPIRING_SOON_30D"
	ExpiryTierCritical ExpiryTier = "EXPIRING_CRITICAL_7D"
	ExpiryTierExpired  ExpiryTier = "EXPIRED"
)

type CertInspectionReport struct {
	SubjectCN          string     `json:"subject_cn"`
	IssuerCN           string     `json:"issuer_cn"`
	SANs               []string   `json:"sans"`
	SignatureAlgorithm string     `json:"signature_algorithm"`
	IsWeakAlgorithm    bool       `json:"is_weak_algorithm"`
	ValidFrom          time.Time  `json:"valid_from"`
	ValidTo            time.Time  `json:"valid_to"`
	DaysRemaining      int        `json:"days_remaining"`
	Tier               ExpiryTier `json:"tier"`
	ChainLength        int        `json:"chain_length"`
	IsChainComplete    bool       `json:"is_chain_complete"`
	SANMatch           bool       `json:"san_match"`
}

type CertInspector struct{}

func NewCertInspector() *CertInspector {
	return &CertInspector{}
}

func (ci *CertInspector) Inspect(chain []*x509.Certificate, expectedHost string) (*CertInspectionReport, error) {
	if len(chain) == 0 {
		return nil, fmt.Errorf("certificate chain is empty")
	}

	leaf := chain[0]
	now := time.Now().UTC()
	daysRem := int(leaf.NotAfter.Sub(now).Hours() / 24)

	tier := ExpiryTierOK
	switch {
	case now.After(leaf.NotAfter):
		tier = ExpiryTierExpired
	case daysRem <= 7:
		tier = ExpiryTierCritical
	case daysRem <= 30:
		tier = ExpiryTierWarning
	}

	sigAlg := leaf.SignatureAlgorithm.String()
	isWeak := strings.Contains(strings.ToLower(sigAlg), "sha1") || strings.Contains(strings.ToLower(sigAlg), "md5")

	// SAN matching including wildcard support
	sanMatch := false
	lowHost := strings.ToLower(expectedHost)
	for _, san := range leaf.DNSNames {
		lowSAN := strings.ToLower(san)
		if lowSAN == lowHost {
			sanMatch = true
			break
		}
		if strings.HasPrefix(lowSAN, "*.") {
			wildSuffix := lowSAN[1:] // e.g. .example.com
			if strings.HasSuffix(lowHost, wildSuffix) && !strings.Contains(strings.TrimSuffix(lowHost, wildSuffix), ".") {
				sanMatch = true
				break
			}
		}
	}

	return &CertInspectionReport{
		SubjectCN:          leaf.Subject.CommonName,
		IssuerCN:           leaf.Issuer.CommonName,
		SANs:               leaf.DNSNames,
		SignatureAlgorithm: sigAlg,
		IsWeakAlgorithm:    isWeak,
		ValidFrom:          leaf.NotBefore,
		ValidTo:            leaf.NotAfter,
		DaysRemaining:      daysRem,
		Tier:               tier,
		ChainLength:        len(chain),
		IsChainComplete:    len(chain) >= 2,
		SANMatch:           sanMatch,
	}, nil
}
