package models

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MinerType define los tipos permitidos de mineros
type MinerType string

const (
	TitularMiner      MinerType = "titular"      // Minero titular 
	SubsistenceMiner  MinerType = "subsistencia" // Minero de subsistencia
)

// Miner representa la entidad del minero en la base de datos.
type Miner struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	gorm.Model   `gorm:"-"` // Deshabilitamos el ID automático de gorm.Model para usar UUID
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// Campos comunes
	FullName    string    `gorm:"not null" json:"full_name"`
	LastName    string    `gorm:"not null" json:"last_name"`
	IDNumber    string    `gorm:"unique;not null" json:"id_number"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `gorm:"unique;not null" json:"email"`
	MinerType   MinerType `gorm:"type:miner_type;not null" json:"miner_type"`

	// Archivos
	IDPhotoFrontPath string `json:"id_photo_front_path"`
	IDPhotoBackPath  string `json:"id_photo_back_path"`
	FacialPhotoPath  string `json:"facial_photo_path"`

	// Documentos específicos
	RuconPath       string `json:"rucon_path,omitempty"`
	OtherDocPath    string `json:"other_doc_path,omitempty"`
	ExploitationContractPath string `json:"exploitation_contract_path,omitempty"`
	EnvironmentalToolPath    string `json:"environmental_tool_path,omitempty"`
	TechnicalToolPath        string `json:"technical_tool_path,omitempty"`
}

// CreateMinerRequest es el DTO para recibir datos de entrada del formulario.
// Los campos de archivo se manejan directamente como *multipart.FileHeader en el controlador.
type CreateMinerRequest struct {
	FullName     string    `form:"full_name" binding:"required"`
	LastName     string    `form:"last_name" binding:"required"`
	IDNumber     string    `form:"id_number" binding:"required"`
	PhoneNumber  string    `form:"phone_number"`
	Email        string    `form:"email" binding:"required,email"`
	MinerType    MinerType `form:"miner_type" binding:"required,oneof=titular subsistencia"`
}


// FileValidationConstraints define las restricciones de validación para cada tipo de documento.
// Las páginas se simulan con un tamaño máximo de archivo (en bytes).
// 1 página ~ 500 KB (aproximación conservadora para PDFs simples).
const (
	Megabyte              int64 = 1024 * 1024
	PhotoMaxSize                = 5 * Megabyte // Fotos de cédula y facial (5MB)
	RuconMaxSize                = 2 * Megabyte // Subsistencia: Rucon (mín 1, máx 2 páginas) -> Max 2 MB
	SubsistenceOtherMaxSize = 10 * Megabyte // Subsistencia: Otros (máx 10 páginas) -> Max 10 MB
	ExploitationContractMaxSize = 15 * Megabyte // Titular: Contrato (25-30 páginas) -> Max 15 MB
	EnvironmentalToolMaxSize = 75 * Megabyte // Titular: Herramienta ambiental (hasta 150 páginas) -> Max 75 MB
	TechnicalToolMaxSize = 50 * Megabyte // Titular: Herramienta técnica (PTO) -> Max 50 MB
)

// DocumentField define la estructura para validar un campo de documento
type DocumentField struct {
	FileHeader  *multipart.FileHeader
	Required    bool
	MaxSizeBytes int64
	AllowedMimeTypes []string
}
