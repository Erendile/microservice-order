package main

type IUserRepository interface {
	Save(CreateUser) error
	FindAll() ([]User, error)
	FindById(string) (User, error)
}
