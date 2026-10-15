package alert

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

type WebhookSigner struct {
	secretKey []byte
}

func NewWebhookSigner(secret string) *WebhookSigner {
	return &WebhookSigner{secretKey: []byte(secret)}
}

func (s *WebhookSigner) SignPayload(payload []byte, timestamp time.Time) (signatureHex string, tsHeader string) {
	tsStr := strconv.FormatInt(timestamp.Unix(), 10)
	toSign := fmt.Sprintf("%s.%s", tsStr, string(payload))

	mac := hmac.New(sha256.New, s.secretKey)
	mac.Write([]byte(toSign))
	sigBytes := mac.Sum(nil)

	return hex.EncodeToString(sigBytes), tsStr
}

func (s *WebhookSigner) VerifySignature(payload []byte, signatureHex, tsHeader string, tolerance time.Duration) bool {
	tsInt, err := strconv.ParseInt(tsHeader, 10, 64)
	if err != nil {
		return false
	}

	sentTime := time.Unix(tsInt, 0)
	if tolerance > 0 && time.Since(sentTime) > tolerance {
		return false
	}

	toSign := fmt.Sprintf("%s.%s", tsHeader, string(payload))
	mac := hmac.New(sha256.New, s.secretKey)
	mac.Write([]byte(toSign))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signatureHex), []byte(expectedSig))
}
