package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	cfg := NewConfiguration()
	redisClient := InitializeRedis(cfg.Redis)

	redisRepository := NewRedisRepository(redisClient)
	jwtService := NewJWTService()
	authService := NewAuthService(redisRepository, jwtService)
	authController := NewAuthController(authService)

	router := mux.NewRouter()
	authController.RegisterRoutes(router)

	log.Printf("Server is running on port %s", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, router); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}
