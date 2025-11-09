package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config almacena todas las variables de entorno para la aplicación.
type Config struct {
	Port          string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	UploadDir     string

	//  Google Cloud Storage
	GCSBucketName           string
	GoogleCredentialsPath   string
}
// LoadConfig carga las variables de entorno desde el archivo .env.
// LoadConfig carga las variables de entorno desde el archivo .env.
func LoadConfig() *Config {
	// Carga el archivo .env si existe
	if err := godotenv.Load(); err != nil {
		log.Println("Advertencia: No se encontró archivo .env. Usando variables de entorno del sistema.")
	}

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "postgres"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),

		// Variables para Google Cloud Storage
		GCSBucketName:         getEnv("GCS_BUCKET_NAME", ""),
		GoogleCredentialsPath: getEnv("GOOGLE_APPLICATION_CREDENTIALS", ""),
	}

	// Crear el directorio local solo si se usa almacenamiento local
	if cfg.UploadDir != "" {
		if _, err := os.Stat(cfg.UploadDir); os.IsNotExist(err) {
			log.Printf("Creando directorio local de carga: %s\n", cfg.UploadDir)
			err = os.MkdirAll(cfg.UploadDir, 0755)
			if err != nil {
				log.Fatalf("Error al crear el directorio de carga: %v", err)
			}
		}
	}

	fmt.Println(" Configuración cargada exitosamente.")
	return cfg
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
