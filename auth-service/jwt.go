package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

type IJWTService interface {
	CreateToken(string, time.Duration) (string, error)
	VerifyToken(string) (*Claims, error)
}

type JWTService struct {
}

func NewJWTService() *JWTService {
	return &JWTService{}
}

var jwtKey = []byte("my_secret_key")

func (s *JWTService) CreateToken(email string, expirationTime time.Duration) (string, error) {
	expiration := time.Now().Add(expirationTime)
	claims := &Claims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiration.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtKey)

	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}

	return signedToken, nil
}

func (s *JWTService) VerifyToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
