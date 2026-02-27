package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	secretKey string
	duration  time.Duration
}

func NewManager(secretKey string, duration time.Duration) *Manager {
	return &Manager{secretKey: secretKey, duration: duration}
}

func (m *Manager) GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(m.duration),
		"iat": time.Now().Unix(), //numericdate
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
func (m *Manager) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.secretKey), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["sub"].(string)
		if !ok {
			return "", jwt.ErrTokenInvalidClaims
		}
		return userID, nil
	}
	return "", jwt.ErrTokenInvalidClaims
}
