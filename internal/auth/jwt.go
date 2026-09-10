package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minimumSecretBytes = 32

type Claims struct {
	DisplayName string `json:"display_name"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

func NewManager(secret, issuer string, ttl time.Duration) (*Manager, error) {
	if len(secret) < minimumSecretBytes {
		return nil, fmt.Errorf("JWT secret must be at least %d bytes", minimumSecretBytes)
	}
	if issuer == "" {
		return nil, errors.New("JWT issuer is required")
	}
	if ttl <= 0 {
		return nil, errors.New("JWT TTL must be positive")
	}
	return &Manager{secret: []byte(secret), issuer: issuer, ttl: ttl}, nil
}

func (m *Manager) Issue(userID, displayName string) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", errors.New("user ID is required")
	}
	now := time.Now().UTC()
	claims := Claims{
		DisplayName: displayName,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID, Issuer: m.issuer,
			IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *Manager) Verify(rawToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		rawToken, &Claims{},
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
			}
			return m.secret, nil
		},
		jwt.WithIssuer(m.issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || strings.TrimSpace(claims.Subject) == "" {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

func ExtractBearer(header string) (string, error) {
	scheme, token, found := strings.Cut(strings.TrimSpace(header), " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return "", errors.New("Bearer token is required")
	}
	return strings.TrimSpace(token), nil
}
