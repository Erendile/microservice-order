package main

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type IUserService interface {
	Create(CreateUser) error
	GetAll() ([]User, error)
	GetById(string) (User, error)
	GetByEmail(string) (User, error)
}

type CreateUser struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
