package repository

import (
	"github.com/batea-fintech/batea-ms-backend/internal/models"
	"gorm.io/gorm"
)

// MinerRepository define la interfaz para la interacción con la base de datos de Mineros.
type MinerRepository interface {
	Create(miner *models.Miner) error
	FindByID(id uint) (*models.Miner, error)
	// Podríamos agregar: FindAll(), Update(), Delete()
}

// minerRepository implementa MinerRepository.
type minerRepository struct {
	db *gorm.DB
}

// NewMinerRepository crea una nueva instancia del repositorio.
func NewMinerRepository(db *gorm.DB) MinerRepository {
	return &minerRepository{db: db}
}

// Create crea un nuevo registro de Minero en la base de datos.
func (r *minerRepository) Create(miner *models.Miner) error {
	return r.db.Create(miner).Error
}

// FindByID busca un Minero por su ID.
func (r *minerRepository) FindByID(id uint) (*models.Miner, error) {
	var miner models.Miner
	if err := r.db.First(&miner, id).Error; err != nil {
		return nil, err
	}
	return &miner, nil
}

