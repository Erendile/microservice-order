package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	cfg := NewConfiguration()

	db, err := NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	router := mux.NewRouter()

	userRepository := NewPostgresRepository(db)
	userService := NewUserService(userRepository)
	userController := NewUserController(userService)
	userController.RegisterRoutes(router)

	log.Printf("Server is running on port %s", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, router); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}
