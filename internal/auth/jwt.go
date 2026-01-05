package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTService handles JWT token operations
type JWTService struct {
	secret               []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secret string, accessDuration, refreshDuration time.Duration) *JWTService {
	return &JWTService{
		secret:               []byte(secret),
		accessTokenDuration:  accessDuration,
		refreshTokenDuration: refreshDuration,
	}
}

// Claims represents JWT claims
type Claims struct {
	AccountID   string   `json:"account_id"`
	CharacterID string   `json:"character_id,omitempty"`
	Role        string   `json:"role"`
	House       string   `json:"house,omitempty"`
	Grade       int      `json:"grade,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	SessionID   string   `json:"session_id"`
	Scopes      []string `json:"scopes"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates an access token
func (s *JWTService) GenerateAccessToken(accountID uuid.UUID, role, sessionID string) (string, error) {
	now := time.Now()
	claims := Claims{
		AccountID: accountID.String(),
		Role:      role,
		SessionID: sessionID,
		Scopes:    []string{"game:play", "chat:read", "chat:write"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "hp-mmo-auth",
			Subject:   accountID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenDuration)),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// GenerateCharacterToken generates a token with character info
func (s *JWTService) GenerateCharacterToken(accountID, characterID uuid.UUID, role, house string, grade int, sessionID string, permissions []string) (string, error) {
	now := time.Now()
	claims := Claims{
		AccountID:   accountID.String(),
		CharacterID: characterID.String(),
		Role:        role,
		House:       house,
		Grade:       grade,
		Permissions: permissions,
		SessionID:   sessionID,
		Scopes:      []string{"game:play", "chat:read", "chat:write"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "hp-mmo-auth",
			Subject:   accountID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenDuration)),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// GenerateRefreshToken generates a refresh token
func (s *JWTService) GenerateRefreshToken(accountID uuid.UUID, sessionID string) (string, error) {
	now := time.Now()
	claims := Claims{
		AccountID: accountID.String(),
		SessionID: sessionID,
		Scopes:    []string{"refresh"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "hp-mmo-auth",
			Subject:   accountID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenDuration)),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken validates and parses a JWT token
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Check expiration
	if claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	// Check issuer
	if claims.Issuer != "hp-mmo-auth" {
		return nil, fmt.Errorf("invalid issuer")
	}

	return claims, nil
}

// ExtractAccountID extracts account ID from token
func (s *JWTService) ExtractAccountID(tokenString string) (uuid.UUID, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(claims.AccountID)
}

// ExtractCharacterID extracts character ID from token
func (s *JWTService) ExtractCharacterID(tokenString string) (uuid.UUID, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	if claims.CharacterID == "" {
		return uuid.Nil, fmt.Errorf("no character ID in token")
	}

	return uuid.Parse(claims.CharacterID)
}
