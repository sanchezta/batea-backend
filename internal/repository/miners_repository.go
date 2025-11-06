package repository

import (
	"errors"

	"github.com/batea-fintech/batea-ms-backend/internal/models"
	"github.com/batea-fintech/batea-ms-backend/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MinerRepository define las operaciones con la base de datos de mineros
type MinerRepository interface {
	Create(miner *models.Miner) error
	FindByEmail(email string) (*models.Miner, error)
	FindByID(id uuid.UUID) (*models.Miner, error)
	Update(miner *models.Miner) error
	Delete(id uuid.UUID) error

	// Método para listar con paginación
	FindAllPaginated(page, limit int) (*utils.Pagination, error)
}

type minerRepository struct {
	db *gorm.DB
}

func NewMinerRepository(db *gorm.DB) MinerRepository {
	return &minerRepository{db}
}

// Create crea un nuevo registro de minero
func (r *minerRepository) Create(miner *models.Miner) error {
	return r.db.Create(miner).Error
}

// FindByEmail busca un minero por email
func (r *minerRepository) FindByEmail(email string) (*models.Miner, error) {
	var miner models.Miner
	err := r.db.Where("email = ?", email).First(&miner).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &miner, nil
}

// FindByID busca un minero por su ID
func (r *minerRepository) FindByID(id uuid.UUID) (*models.Miner, error) {
	var miner models.Miner
	err := r.db.First(&miner, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &miner, nil
}

// Update actualiza los datos del minero
func (r *minerRepository) Update(miner *models.Miner) error {
	return r.db.Save(miner).Error
}

// Delete elimina un minero (soft delete)
func (r *minerRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Miner{}, id).Error
}

func (r *minerRepository) FindAllPaginated(page, limit int) (*utils.Pagination, error) {
	var miners []models.Miner
	return utils.Paginate(r.db, &models.Miner{}, page, limit, &miners)
}
