package auth

import (
    "store/internal/config"
	"context"
	"errors"
    "strconv"
	"store/internal/models"
    "store/internal/util"
	"store/internal/repositories/auth"
	"golang.org/x/crypto/bcrypt"

)

// Business Logic

type AuthService interface {
	Register(ctx context.Context, user models.User) (models.User, error)
	Login(ctx context.Context, email string, password string) (string, error)
}

type AuthServiceImpl struct {
	repo auth.AuthRepository
    config *config.ApplicationConfig
}

func NewAuthService(repo auth.AuthRepository, config *config.ApplicationConfig) AuthService {
    return &AuthServiceImpl{repo: repo, config: config}
}

// methods 


func (s *AuthServiceImpl) Register(ctx context.Context, user models.User) (models.User, error) {

	if user.Email == "" || user.Password == "" {
		return models.User{}, errors.New("email and password required")
	}

	if user.FirstName == ""  || user.LastName == ""{
		return models.User{}, errors.New("First Name and Last Name cannot be Empty")
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	user.Password = string(hashedPassword)

	createdUser, err := s.repo.Register(ctx, user)
	if err != nil {
		return models.User{}, err
	}
    // Create Stripe Customer Id


	createdUser.Password = ""

	return createdUser, nil
}

func (s *AuthServiceImpl) Login(ctx context.Context, email string, password string) (string, error) {

	if email == "" || password == "" {
		return "", errors.New("email and password required")
	}

	user, err := s.repo.Login(ctx, email, password)
	if err != nil {
		return "", errors.New("invalid credentials - service")
	}

	token, err := util.GenerateAuthToken(user.Email, strconv.FormatInt(user.ID, 10), s.config)
	if err != nil {
		return "", err
	}

	return token, nil
}
