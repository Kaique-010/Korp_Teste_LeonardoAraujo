package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret    []byte
	expiresIn time.Duration
}

type Claims struct {
	UserID uint   `json:"sub"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func ValidateJWTSecret(secret []byte) error {
	if len(secret) == 0 {
		return errors.New("JWT_SECRET não pode ser vazio")
	}
	if len(secret) < 16 {
		return errors.New("JWT_SECRET deve ter pelo menos 16 caracteres")
	}
	return nil
}

func NewJWTService(secret string, expiresIn time.Duration) *JWTService {
	return &JWTService{
		secret:    []byte(secret),
		expiresIn: expiresIn,
	}
}

func (s *JWTService) Generate(
	userID uint,
	email string,
) (string, error) {

	now := time.Now()

	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(s.secret)
}

func (s *JWTService) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("método de assinatura inválido")
			}
			return s.secret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("token inválido")
	}

	return claims, nil
}
