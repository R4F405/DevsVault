package application

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	authdomain "github.com/devsvault/devsvault/apps/api/internal/auth/domain"
)

var ErrInvalidToken = errors.New("invalid token")

type HMACTokenIssuer struct {
	key []byte
	ttl time.Duration
}

type tokenClaims struct {
	Subject string   `json:"sub"`
	Type    string   `json:"typ"`
	Roles   []string `json:"roles"`
	Expiry  int64    `json:"exp"`
}

func NewHMACTokenIssuer(key []byte, ttl time.Duration) *HMACTokenIssuer {
	return &HMACTokenIssuer{key: key, ttl: ttl}
}

func (i *HMACTokenIssuer) Issue(actor authdomain.Actor) (IssuedToken, error) {
	expiresAt := time.Now().UTC().Add(i.ttl)
	claims := tokenClaims{Subject: actor.ID, Type: string(actor.Type), Roles: actor.Roles, Expiry: expiresAt.Unix()}
	payload, err := json.Marshal(claims)
	if err != nil {
		return IssuedToken{}, err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := i.sign(encodedPayload)
	return IssuedToken{AccessToken: encodedPayload + "." + signature, ExpiresAt: expiresAt}, nil
}

func (i *HMACTokenIssuer) Verify(token string) (authdomain.Actor, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return authdomain.Anonymous(), ErrInvalidToken
	}
	if !hmac.Equal([]byte(i.sign(parts[0])), []byte(parts[1])) {
		return authdomain.Anonymous(), ErrInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return authdomain.Anonymous(), ErrInvalidToken
	}
	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return authdomain.Anonymous(), ErrInvalidToken
	}
	if time.Now().UTC().Unix() > claims.Expiry {
		return authdomain.Anonymous(), ErrInvalidToken
	}
	return authdomain.Actor{ID: claims.Subject, Type: authdomain.ActorType(claims.Type), Roles: claims.Roles}, nil
}

func (i *HMACTokenIssuer) sign(payload string) string {
	mac := hmac.New(sha256.New, i.key)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
