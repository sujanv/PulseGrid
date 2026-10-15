package checker

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

type MTLSProfile struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	ClientCertPEM      string `json:"client_cert_pem"`
	ClientKeyPEM       string `json:"client_key_pem"`
	RootCAPEM          string `json:"root_ca_pem,omitempty"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
}

type MTLSConfigLoader struct{}

func NewMTLSConfigLoader() *MTLSConfigLoader {
	return &MTLSConfigLoader{}
}

func (l *MTLSConfigLoader) BuildTLSConfig(profile *MTLSProfile) (*tls.Config, error) {
	if profile == nil {
		return &tls.Config{}, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: profile.InsecureSkipVerify,
	}

	if profile.ClientCertPEM != "" && profile.ClientKeyPEM != "" {
		cert, err := tls.X509KeyPair([]byte(profile.ClientCertPEM), []byte(profile.ClientKeyPEM))
		if err != nil {
			return nil, fmt.Errorf("failed to parse client x509 key pair: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if profile.RootCAPEM != "" {
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM([]byte(profile.RootCAPEM)) {
			return nil, fmt.Errorf("failed to append root CA from PEM")
		}
		tlsConfig.RootCAs = caPool
	}

	return tlsConfig, nil
}
