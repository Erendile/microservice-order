package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

type AuthController struct {
	authService IAuthService
}

func NewAuthController(authService IAuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (c *AuthController) register(w http.ResponseWriter, r *http.Request) {
	var creds RegisterCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	tokens, err := c.authService.Register(creds, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func (c *AuthController) login(w http.ResponseWriter, r *http.Request) {
	var creds LoginCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	tokens, err := c.authService.Login(creds, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func (c *AuthController) refresh(w http.ResponseWriter, r *http.Request) {
	tokenReq, err := c.getTokens(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tokens, err := c.authService.Refresh(tokenReq, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func (c *AuthController) logout(w http.ResponseWriter, r *http.Request) {
	tokenReq, err := c.getTokens(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = c.authService.Logout(tokenReq, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *AuthController) getTokens(r *http.Request) (Tokens, error) {
	var tokenReq Tokens

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return tokenReq, fmt.Errorf("Missing or invalid Authorization header")
	}
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		return tokenReq, fmt.Errorf("Missing refresh token cookie")
	}
	refreshToken := cookie.Value

	tokenReq.AccessToken = accessToken
	tokenReq.RefreshToken = refreshToken

	return tokenReq, nil
}

func (c *AuthController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/auth/register", c.register).Methods("POST")
	router.HandleFunc("/auth/login", c.login).Methods("POST")
	router.HandleFunc("/auth/refresh", c.refresh).Methods("POST")
	router.HandleFunc("/auth/logout", c.logout).Methods("POST")
}
