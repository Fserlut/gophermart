package user

import (
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Fserlut/gophermart/internal/lib"
	"github.com/Fserlut/gophermart/internal/models/user"
)

type ServiceUser struct {
	userRepository userRepository
}

type userRepository interface {
	CreateUser(user.User) error
	GetUserByLogin(string) (*user.User, error)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func verifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (u ServiceUser) Register(userCreate user.RegisterOrLoginRequest) (*http.Cookie, error) {
	hashPass, err := hashPassword(userCreate.Password)
	if err != nil {
		return nil, err
	}

	user := user.User{
		UUID:     uuid.New().String(),
		Login:    userCreate.Login,
		Password: hashPass,
	}

	err = u.userRepository.CreateUser(user)

	if err != nil {
		return nil, err
	}

	authCookie, _ := lib.GenerateAuthCookie(user.UUID)

	return authCookie, nil
}

func (u ServiceUser) Login(userCreate user.RegisterOrLoginRequest) (*http.Cookie, error) {
	user, err := u.userRepository.GetUserByLogin(userCreate.Login)
	if err != nil {
		return nil, err
	}

	if verifyPassword(user.Password, userCreate.Password) {
		authCookie, _ := lib.GenerateAuthCookie(user.UUID)

		return authCookie, nil
	}
	return nil, &lib.ErrWrongPasswordOrLogin{}
}

func NewUserService(userRepository userRepository) *ServiceUser {
	return &ServiceUser{
		userRepository: userRepository,
	}
}
