package repository

import (
	"errors"

	"github.com/sanchezta/batea-backend/internal/models"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("usuario no encontrado")

type UserRepository interface {
	Create(user *models.User) error
	FindByPhone(phone string) (*models.User, error)
	FindByID(id string) (*models.User, error)
	FindByFirebaseUID(uid string) (*models.User, error)
	UpdateVerificationStatus(id string, verified bool) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByPhone(phone string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("phone_number = ?", phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByFirebaseUID(uid string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("firebase_uid = ?", uid).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateVerificationStatus(id string, verified bool) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).
		Update("is_verified", verified).Error
}
