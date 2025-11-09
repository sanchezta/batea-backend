package service

import (
	"errors"

	"github.com/sanchezta/batea-backend/internal/models"
	"github.com/sanchezta/batea-backend/internal/repository"
	"github.com/sanchezta/batea-backend/internal/utils"
)

var (
	ErrUserAlreadyExists = errors.New("ya existe un usuario con este número de teléfono")
	ErrPasswordHash      = errors.New("error al encriptar la contraseña")
)

type UserService interface {
	RegisterUser(req *models.UserRegisterRequest) (*models.UserResponse, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo}
}

func (s *userService) RegisterUser(req *models.UserRegisterRequest) (*models.UserResponse, error) {
	existing, err := s.userRepo.FindByPhone(req.PhoneNumber)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	var hash string
	if req.FirebaseUID == "" {
		hash, err = utils.HashPassword(req.Password)
		if err != nil {
			return nil, ErrPasswordHash
		}
	}

	user := &models.User{
		PhoneNumber:  req.PhoneNumber,
		PasswordHash: hash,
		IsVerified:   req.IsVerified,
		FirebaseUID:  req.FirebaseUID, // Para cuando uses Identity Platform
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return &models.UserResponse{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		IsVerified:  user.IsVerified,
		CreatedAt:   user.CreatedAt,
	}, nil
}
