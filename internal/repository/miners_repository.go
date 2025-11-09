package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sanchezta/batea-backend/internal/models"
	"github.com/sanchezta/batea-backend/internal/utils"
	"gorm.io/gorm"
)

var ErrMinerNotFound = errors.New("minero no encontrado")

type MinerRepository interface {
	Create(miner *models.Miner) error
	FindByEmail(email string) (*models.Miner, error)
	FindByID(id uuid.UUID) (*models.Miner, error)
	FindByUserID(userID uuid.UUID) (*models.Miner, error)
	FindByFirebaseUID(uid string) (*models.Miner, error)
	Update(miner *models.Miner) error
	Delete(id uuid.UUID) error
	FindAllPaginated(page, limit int) (*utils.Pagination, error)
}

type minerRepository struct {
	db *gorm.DB
}

func NewMinerRepository(db *gorm.DB) MinerRepository {
	return &minerRepository{db}
}

func (r *minerRepository) Create(miner *models.Miner) error {
	return r.db.Create(miner).Error
}

func (r *minerRepository) FindByEmail(email string) (*models.Miner, error) {
	var miner models.Miner
	if err := r.db.Where("email = ?", email).First(&miner).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMinerNotFound
		}
		return nil, err
	}
	return &miner, nil
}

func (r *minerRepository) FindByID(id uuid.UUID) (*models.Miner, error) {
	var miner models.Miner
	if err := r.db.First(&miner, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMinerNotFound
		}
		return nil, err
	}
	return &miner, nil
}

func (r *minerRepository) FindByUserID(userID uuid.UUID) (*models.Miner, error) {
	var miner models.Miner
	if err := r.db.Where("user_id = ?", userID).First(&miner).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMinerNotFound
		}
		return nil, err
	}
	return &miner, nil
}

func (r *minerRepository) FindByFirebaseUID(uid string) (*models.Miner, error) {
	var miner models.Miner
	err := r.db.Joins("JOIN users ON users.id = miners.user_id").
		Where("users.firebase_uid = ?", uid).
		First(&miner).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMinerNotFound
		}
		return nil, err
	}
	return &miner, nil
}

func (r *minerRepository) Update(miner *models.Miner) error {
	return r.db.Save(miner).Error
}

func (r *minerRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Miner{}, id).Error
}

func (r *minerRepository) FindAllPaginated(page, limit int) (*utils.Pagination, error) {
	var miners []models.Miner
	return utils.Paginate(r.db, &models.Miner{}, page, limit, &miners)
}
