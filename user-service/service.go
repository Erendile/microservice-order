package main

type UserService struct {
	userRepository IUserRepository
}

func NewUserService(userRepository IUserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func (us *UserService) Create(user CreateUser) error {
	return us.userRepository.Save(user)
}

func (us *UserService) GetAll() ([]User, error) {
	return us.userRepository.FindAll()
}

func (us *UserService) GetById(id string) (User, error) {
	return us.userRepository.FindById(id)
}
