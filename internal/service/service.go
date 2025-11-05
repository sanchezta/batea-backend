package service

import (
	"fmt"
	"log"
	"mime/multipart"
	"strings"

	"github.com/batea-fintech/batea-ms-backend/internal/config"
	"github.com/batea-fintech/batea-ms-backend/internal/models"
	"github.com/batea-fintech/batea-ms-backend/internal/repository"
	"github.com/batea-fintech/batea-ms-backend/internal/utils"
)

// MinerService define la interfaz para la lógica de negocio de Mineros.
type MinerService interface {
	CreateMiner(
		req *models.CreateMinerRequest,
		files map[string]*multipart.FileHeader,
	) (*models.Miner, error)
	// Podríamos agregar: GetMinerByID(), ValidateMinerData(), etc.
}

// minerService implementa MinerService.
type minerService struct {
	repo repository.MinerRepository
	cfg *config.Config
}

// NewMinerService crea una nueva instancia del servicio.
func NewMinerService(repo repository.MinerRepository, cfg *config.Config) MinerService {
	return &minerService{repo: repo, cfg: cfg}
}

// CreateMiner maneja toda la lógica para registrar un nuevo minero, incluyendo la validación y el almacenamiento de archivos.
func (s *minerService) CreateMiner(
	req *models.CreateMinerRequest,
	files map[string]*multipart.FileHeader,
) (*models.Miner, error) {
	
	// 1. Validar los archivos antes de la persistencia
	validationError := s.validateMinerFiles(req.MinerType, files)
	if validationError != nil {
		return nil, validationError
	}

	// 2. Crear el objeto Miner e inicializarlo con los datos de la solicitud
	miner := &models.Miner{
		FullName:    req.FullName,
		LastName:    req.LastName,
		IDNumber:    req.IDNumber,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		MinerType:   req.MinerType,
	}

	// 3. Procesar y guardar archivos
	// Nota: En un entorno de producción, aquí se subiría a GCP (GCS) y se guardaría la URL pública.
	// Aquí simulamos guardando localmente y almacenando la ruta local.
	
	var filesToSave = map[string]string{
		"id_photo_front": "", 
		"id_photo_back": "", 
		"facial_photo": "",
		"rucon": "", 
		"other_doc": "",
		"exploitation_contract": "",
		"environmental_tool": "",
		"technical_tool": "",
	}
	
	// Archivos comunes
	commonFiles := map[string]string{
		"id_photo_front": "cedulas", 
		"id_photo_back": "cedulas", 
		"facial_photo": "facial",
	}

	for field, subdir := range commonFiles {
		if file, ok := files[field]; ok {
			path, err := utils.SaveFile(s.cfg, file, subdir)
			if err != nil {
				return nil, fmt.Errorf("fallo al guardar archivo común %s: %w", field, err)
			}
			filesToSave[field] = path
		}
	}

	// Archivos específicos según el tipo de minero
	switch req.MinerType {
	case models.SubsistenceMiner:
		if file, ok := files["rucon"]; ok {
			path, err := utils.SaveFile(s.cfg, file, "subsistencia/rucon")
			if err != nil {
				return nil, fmt.Errorf("fallo al guardar RUCON: %w", err)
			}
			filesToSave["rucon"] = path
		}
		if file, ok := files["other_doc"]; ok { // Opcional
			path, err := utils.SaveFile(s.cfg, file, "subsistencia/otros")
			if err != nil {
				return nil, fmt.Errorf("fallo al guardar otro documento: %w", err)
			}
			filesToSave["other_doc"] = path
		}

	case models.TitularMiner:
		specificFiles := map[string]string{
			"exploitation_contract": "titular/contrato",
			"environmental_tool": "titular/ambiental",
			"technical_tool": "titular/tecnica",
		}
		for field, subdir := range specificFiles {
			if file, ok := files[field]; ok {
				path, err := utils.SaveFile(s.cfg, file, subdir)
				if err != nil {
					return nil, fmt.Errorf("fallo al guardar archivo titular %s: %w", field, err)
				}
				filesToSave[field] = path
			}
		}
	}
	
	// 4. Asignar rutas al modelo (Clean Architecture: Mapeo DTO a Entidad)
	miner.IDPhotoFrontPath = filesToSave["id_photo_front"]
	miner.IDPhotoBackPath = filesToSave["id_photo_back"]
	miner.FacialPhotoPath = filesToSave["facial_photo"]
	miner.RuconPath = filesToSave["rucon"]
	miner.OtherDocPath = filesToSave["other_doc"]
	miner.ExploitationContractPath = filesToSave["exploitation_contract"]
	miner.EnvironmentalToolPath = filesToSave["environmental_tool"]
	miner.TechnicalToolPath = filesToSave["technical_tool"]

	// 5. Persistencia en base de datos
	if err := s.repo.Create(miner); err != nil {
		// En caso de error de DB, se podría agregar lógica para limpiar archivos subidos
		log.Printf("Error de DB. Limpieza de archivos omitida por simplicidad: %v", err)
		if strings.Contains(err.Error(), "unique_violation") {
			return nil, fmt.Errorf("ya existe un minero registrado con esta cédula o correo electrónico")
		}
		return nil, fmt.Errorf("fallo al guardar el minero en la base de datos: %w", err)
	}

	return miner, nil
}

// validateMinerFiles valida la presencia, tamaño y tipo de archivos según el MinerType.
func (s *minerService) validateMinerFiles(minerType models.MinerType, files map[string]*multipart.FileHeader) error {
	
	// Documentos comunes (Requeridos)
	baseFields := map[string]models.DocumentField{
		"id_photo_front": {
			FileHeader: files["id_photo_front"],
			Required: true,
			MaxSizeBytes: models.PhotoMaxSize,
			AllowedMimeTypes: []string{"image/jpeg", "image/png"},
		},
		"id_photo_back": {
			FileHeader: files["id_photo_back"],
			Required: true,
			MaxSizeBytes: models.PhotoMaxSize,
			AllowedMimeTypes: []string{"image/jpeg", "image/png"},
		},
		"facial_photo": {
			FileHeader: files["facial_photo"],
			Required: true,
			MaxSizeBytes: models.PhotoMaxSize,
			AllowedMimeTypes: []string{"image/jpeg", "image/png"},
		},
	}
	
	// Validar campos comunes
	for fieldName, field := range baseFields {
		if err := utils.ValidateFile(field); err != nil {
			return fmt.Errorf("error de validación en el archivo '%s': %w", fieldName, err)
		}
	}

	// Documentos específicos
	var specificFields map[string]models.DocumentField
	
	pdfMimes := []string{"application/pdf", "pdf"}

	switch minerType {
	case models.SubsistenceMiner:
		specificFields = map[string]models.DocumentField{
			"rucon": {
				FileHeader: files["rucon"],
				Required: true,
				MaxSizeBytes: models.RuconMaxSize,
				AllowedMimeTypes: pdfMimes,
			},
			"other_doc": { // Opcional
				FileHeader: files["other_doc"],
				Required: false, 
				MaxSizeBytes: models.SubsistenceOtherMaxSize,
				AllowedMimeTypes: pdfMimes,
			},
		}

	case models.TitularMiner:
		specificFields = map[string]models.DocumentField{
			"exploitation_contract": {
				FileHeader: files["exploitation_contract"],
				Required: true,
				MaxSizeBytes: models.ExploitationContractMaxSize,
				AllowedMimeTypes: pdfMimes,
			},
			"environmental_tool": {
				FileHeader: files["environmental_tool"],
				Required: true,
				MaxSizeBytes: models.EnvironmentalToolMaxSize,
				AllowedMimeTypes: pdfMimes,
			},
			"technical_tool": {
				FileHeader: files["technical_tool"],
				Required: true,
				MaxSizeBytes: models.TechnicalToolMaxSize,
				AllowedMimeTypes: pdfMimes,
			},
		}
	default:
		return fmt.Errorf("tipo de minero no válido")
	}

	// Validar campos específicos
	for fieldName, field := range specificFields {
		if err := utils.ValidateFile(field); err != nil {
			return fmt.Errorf("error de validación en el archivo específico '%s': %w", fieldName, err)
		}
	}

	return nil
}
