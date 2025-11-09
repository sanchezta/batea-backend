package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	PhoneNumber  string         `gorm:"unique;not null" json:"phone_number"`
	PasswordHash string         `gorm:"not null" json:"-"`

	// Preparado para Identity Platform (opcional / nullable)
	FirebaseUID  string         `gorm:"uniqueIndex;default:null" json:"firebase_uid,omitempty"`

	// Estado de verificación (por SMS/Email)
	IsVerified   bool           `gorm:"default:false" json:"is_verified"`
}

// DTO de entrada para registrar usuario
type UserRegisterRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,e164"`
	// Para flujo local con contraseña
	Password    string `json:"password,omitempty" binding:"omitempty,min=6,max=32"`
	// Para flujo con Identity Platform (opcional)
	FirebaseUID string `json:"firebase_uid,omitempty"`
	// Permite marcar verificado si viene de un proveedor externo
	IsVerified  bool   `json:"is_verified,omitempty"`
}

// DTO de salida
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	IsVerified  bool      `json:"is_verified"`
	CreatedAt   time.Time `json:"created_at"`
}
