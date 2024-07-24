package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type RegisterCredentials struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type IAuthService interface {
	Register(RegisterCredentials, http.ResponseWriter) (*Tokens, error)
	Login(LoginCredentials, http.ResponseWriter) (*Tokens, error)
	Refresh(Tokens, http.ResponseWriter) (*Tokens, error)
	Logout(Tokens, http.ResponseWriter) error
}

type AuthService struct {
	redisRepository IRedisRepository
	jwtService      IJWTService
}

func NewAuthService(redisRepository IRedisRepository, jwtService IJWTService) *AuthService {
	return &AuthService{
		redisRepository: redisRepository,
		jwtService:      jwtService,
	}
}

func (s *AuthService) Register(creds RegisterCredentials, w http.ResponseWriter) (*Tokens, error) {
	if err := s.createUser(creds); err != nil {
		return nil, err
	}

	return s.createAndSetTokens(creds.Email, w)
}

func (s *AuthService) Login(creds LoginCredentials, w http.ResponseWriter) (*Tokens, error) {
	user, err := s.getUserByEmail(creds.Email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return s.createAndSetTokens(creds.Email, w)
}

func (s *AuthService) Refresh(tokenReq Tokens, w http.ResponseWriter) (*Tokens, error) {
	email, err := s.redisRepository.GetToken(tokenReq.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}
	s.redisRepository.DeleteToken(tokenReq.RefreshToken)
	return s.createAndSetTokens(email, w)
}

func (s *AuthService) Logout(tokenReq Tokens, w http.ResponseWriter) error {
	s.redisRepository.DeleteToken(tokenReq.RefreshToken)
	s.deleteRefreshTokenCookie(w)
	return nil
}

func (s *AuthService) createUser(creds RegisterCredentials) error {
	userServiceURL := "http://localhost:8080/users"
	jsonData, _ := json.Marshal(creds)
	resp, err := http.Post(userServiceURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil || resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to register user:" + err.Error())
	}
	return nil
}

func (s *AuthService) createAndSetTokens(email string, w http.ResponseWriter) (*Tokens, error) {
	accessToken, err := s.jwtService.CreateToken(email, time.Minute*15)
	if err != nil {
		return nil, fmt.Errorf("error creating access token")
	}

	refreshToken, err := s.jwtService.CreateToken(email, time.Hour*24*7)
	if err != nil {
		return nil, fmt.Errorf("error creating refresh token")
	}

	if err := s.redisRepository.SetToken(refreshToken, email, time.Hour*24*7); err != nil {
		return nil, fmt.Errorf("error saving refresh token")
	}

	s.setRefreshTokenCookie(w, refreshToken, time.Hour*24*7)

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) getUserByEmail(email string) (*User, error) {
	userServiceURL := "http://localhost:8080/users/email/" + email
	resp, err := http.Get(userServiceURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user not found")
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("error decoding response")
	}

	return &user, nil
}

func (s *AuthService) setRefreshTokenCookie(w http.ResponseWriter, token string, expirationTime time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(expirationTime),
	})
}

func (s *AuthService) deleteRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
