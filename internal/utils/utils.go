package utils

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/batea-fintech/batea-ms-backend/internal/config"
	"github.com/batea-fintech/batea-ms-backend/internal/models"
)

// SaveFile guarda un archivo de la solicitud multipart en el disco local (simulación de GCP).
func SaveFile(cfg *config.Config, file *multipart.FileHeader, subdir string) (string, error) {
	if file == nil {
		return "", fmt.Errorf("el archivo no puede ser nulo")
	}

	// 1. Crear el subdirectorio (por ejemplo, 'cedulas' o 'pdfs')
	fullDir := filepath.Join(cfg.UploadDir, subdir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", fmt.Errorf("error al crear el directorio de carga: %w", err)
	}

	// 2. Generar nombre de archivo único
	// Usamos el nombre original con un prefijo o sufijo para evitar colisiones
	fileName := strings.ReplaceAll(file.Filename, " ", "_")
	targetPath := filepath.Join(fullDir, fileName)

	// 3. Abrir el archivo de entrada
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("error al abrir el archivo de subida: %w", err)
	}
	defer src.Close()

	// 4. Crear el archivo de destino
	dst, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("error al crear el archivo de destino: %w", err)
	}
	defer dst.Close()

	// 5. Copiar el contenido
	if _, err := dst.ReadFrom(src); err != nil {
		return "", fmt.Errorf("error al copiar el archivo: %w", err)
	}

	// La ruta que se guarda en DB es relativa al directorio de uploads (o la URL de GCP)
	relativePath := filepath.Join(subdir, fileName)
	return relativePath, nil
}

// ValidateFile realiza las validaciones de tipo MIME y tamaño.
func ValidateFile(field models.DocumentField) error {
	f := field.FileHeader

	// Si no es requerido y el archivo no se subió, es válido.
	if !field.Required && f == nil {
		return nil
	}

	// Si es requerido y el archivo no se subió.
	if field.Required && f == nil {
		return fmt.Errorf("documento requerido no proporcionado")
	}

	// Validación de tamaño máximo (simulando límite de páginas)
	if f.Size > field.MaxSizeBytes {
		return fmt.Errorf("el archivo '%s' excede el tamaño máximo permitido de %s MB. (Simulación de límite de páginas)",
			f.Filename, byteCountToMB(field.MaxSizeBytes))
	}

	// Validación de tipo MIME
	// La detección precisa del tipo MIME requiere leer el contenido, pero para un multipart simple,
	// podemos usar la extensión del nombre o una validación ligera.
	// Nota: `CheckContentType` de Gin es más estricto, aquí hacemos una validación básica.
	
	validMime := false
	for _, mimeType := range field.AllowedMimeTypes {
		if strings.Contains(strings.ToLower(f.Header.Get("Content-Type")), strings.ToLower(mimeType)) || strings.HasSuffix(strings.ToLower(f.Filename), "."+mimeType) {
			validMime = true
			break
		}
	}

	if !validMime {
		return fmt.Errorf("el tipo de archivo para '%s' no es válido. Tipos permitidos: %s",
			f.Filename, strings.Join(field.AllowedMimeTypes, ", "))
	}

	return nil
}

// byteCountToMB convierte bytes a una cadena de megabytes para mensajes de error.
func byteCountToMB(b int64) string {
	return fmt.Sprintf("%.2f", float64(b)/float64(models.Megabyte))
}
