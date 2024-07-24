package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

type UserController struct {
	userService IUserService
}

func NewUserController(userService IUserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (u *UserController) create(w http.ResponseWriter, r *http.Request) {
	var user CreateUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := u.userService.Create(user); err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (u *UserController) getAll(w http.ResponseWriter, r *http.Request) {
	users, err := u.userService.GetAll()
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (u *UserController) getById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	user, err := u.userService.GetById(id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (u *UserController) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/users", u.create).Methods("POST")
	r.HandleFunc("/users", u.getAll).Methods("GET")
	r.HandleFunc("/users/{id}", u.getById).Methods("GET")
}
