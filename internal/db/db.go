package db

import (
	"fmt"
	"log"
	"time"

	"github.com/batea-fintech/batea-ms-backend/internal/config"
	"github.com/batea-fintech/batea-ms-backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitPostgres inicializa la conexión a PostgreSQL y realiza las migraciones.
func InitPostgres(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Santiago",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	var db *gorm.DB
	var err error

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Intento %d/%d: Fallo al conectar con PostgreSQL. Reintentando en 5 segundos...", i+1, maxRetries)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("fallo fatal al conectar con la base de datos: %w", err)
	}

	// 🟢 Paso nuevo: crear tipo ENUM miner_type si no existe
	createMinerTypeEnum(db)

	// Migrar modelos
	if err := db.AutoMigrate(&models.Miner{}); err != nil {
		return nil, fmt.Errorf("fallo en la migración de la base de datos: %w", err)
	}

	log.Println("Conexión a PostgreSQL y migración completadas con éxito.")
	return db, nil
}

// 🧩 Función auxiliar para crear el tipo ENUM si no existe
func createMinerTypeEnum(db *gorm.DB) {
	sql := `
	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'miner_type') THEN
			CREATE TYPE miner_type AS ENUM ('titular', 'subsistencia');
		END IF;
	END$$;
	`
	if err := db.Exec(sql).Error; err != nil {
		log.Fatalf("Error al crear el tipo ENUM miner_type: %v", err)
	}
	log.Println("Tipo ENUM 'miner_type' verificado o creado correctamente.")
}
