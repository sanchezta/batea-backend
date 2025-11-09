package service

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/sanchezta/batea-backend/internal/config"
	"github.com/sanchezta/batea-backend/internal/models"
	"github.com/sanchezta/batea-backend/internal/repository"
	"github.com/sanchezta/batea-backend/internal/utils"
)

// MinerService define la interfaz para la lógica de negocio de Mineros.
// 👇 Recibe userID como primer parámetro (NO usamos req.UserID)
type MinerService interface {
	CreateMiner(
		userID uuid.UUID,
		req *models.CreateMinerRequest,
		files map[string]*multipart.FileHeader,
	) (*models.Miner, string, string, error)

	GetAllMiners(page, limit int) (*utils.Pagination, error)
	GenerateTOTP(minerID uuid.UUID) (string, error)
	ValidateTOTP(minerID uuid.UUID, code string) (bool, error)
}

type minerService struct {
	repo     repository.MinerRepository
	userRepo repository.UserRepository // para verificar existencia del usuario
	cfg      *config.Config
}

// NewMinerService crea una nueva instancia del servicio de mineros.
// Si no quieres validar usuario, puedes pasar nil en userRepo y saltar esa verificación.
func NewMinerService(repo repository.MinerRepository, userRepo repository.UserRepository, cfg *config.Config) MinerService {
	return &minerService{
		repo:     repo,
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// CreateMiner: ahora recibe userID (no depende de req.UserID).
func (s *minerService) CreateMiner(
	userID uuid.UUID,
	req *models.CreateMinerRequest,
	files map[string]*multipart.FileHeader,
) (*models.Miner, string, string, error) {

	// Validar userID
	if userID == uuid.Nil {
		return nil, "", "", errors.New("falta el ID del usuario asociado al minero")
	}

	// (Opcional) Verificar que el usuario exista si userRepo no es nil
	if s.userRepo != nil {
		user, err := s.userRepo.FindByID(userID.String())
		if err != nil {
			return nil, "", "", fmt.Errorf("error al buscar usuario asociado: %w", err)
		}
		if user == nil {
			return nil, "", "", errors.New("el usuario asociado no existe o fue eliminado")
		}
	}

	// Validar archivos antes de persistencia
	if err := s.validateMinerFiles(req.MinerType, files); err != nil {
		return nil, "", "", err
	}

	// Crear el objeto Miner vinculado al usuario
	miner := &models.Miner{
		UserID:      userID, // vínculo real con User
		FullName:    req.FullName,
		LastName:    req.LastName,
		// PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		MinerType:   req.MinerType,
	}

	// Map de rutas a guardar
	filesToSave := map[string]string{
		"id_photo_front":        "",
		"id_photo_back":         "",
		"facial_photo":          "",
		"rucon":                 "",
		"other_doc":             "",
		"exploitation_contract": "",
		"environmental_tool":    "",
		"technical_tool":        "",
	}

	// Comunes
	commonFiles := map[string]string{
		"id_photo_front": "cedulas",
		"id_photo_back":  "cedulas",
		"facial_photo":   "facial",
	}
	for field, subdir := range commonFiles {
		if file, ok := files[field]; ok {
			path, err := utils.SaveFile(s.cfg, file, subdir)
			if err != nil {
				return nil, "", "", fmt.Errorf("fallo al guardar archivo %s: %w", field, err)
			}
			filesToSave[field] = path
		}
	}

	// Específicos por tipo
	switch req.MinerType {
	case models.SubsistenceMiner:
		if file, ok := files["rucon"]; ok {
			path, err := utils.SaveFile(s.cfg, file, "subsistencia/rucon")
			if err != nil {
				return nil, "", "", fmt.Errorf("fallo al guardar RUCON: %w", err)
			}
			filesToSave["rucon"] = path
		}
		
		if file, ok := files["other_doc"]; ok && file != nil {
    path, err := utils.SaveFile(s.cfg, file, "subsistencia/otros")
    if err != nil {
        return nil, "", "", fmt.Errorf("fallo al guardar otro documento: %w", err)
    }
    filesToSave["other_doc"] = path
}
	case models.TitularMiner:
		spec := map[string]string{
			"exploitation_contract": "titular/contrato",
			"environmental_tool":    "titular/ambiental",
			"technical_tool":        "titular/tecnica",
		}
		for field, subdir := range spec {
			if file, ok := files[field]; ok {
				path, err := utils.SaveFile(s.cfg, file, subdir)
				if err != nil {
					return nil, "", "", fmt.Errorf("fallo al guardar archivo titular %s: %w", field, err)
				}
				filesToSave[field] = path
			}
		}
	default:
		return nil, "", "", fmt.Errorf("tipo de minero '%s' no reconocido", req.MinerType)
	}

	// Asignar rutas
	miner.IDPhotoFrontPath = filesToSave["id_photo_front"]
	miner.IDPhotoBackPath = filesToSave["id_photo_back"]
	miner.FacialPhotoPath = filesToSave["facial_photo"]
	miner.RuconPath = filesToSave["rucon"]
	miner.OtherDocPath = filesToSave["other_doc"]
	miner.ExploitationContractPath = filesToSave["exploitation_contract"]
	miner.EnvironmentalToolPath = filesToSave["environmental_tool"]
	miner.TechnicalToolPath = filesToSave["technical_tool"]

	// TOTP
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Batea Fintech",
		AccountName: req.Email,
		Period:      30,
		Digits:      otp.DigitsSix,
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("error generando TOTP: %w", err)
	}
	miner.TOTPSecret = key.Secret()
	qrURL := key.URL()

	code, err := totp.GenerateCodeCustom(miner.TOTPSecret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, "", "", fmt.Errorf("error generando código TOTP: %w", err)
	}

	// Persistir
	if err := s.repo.Create(miner); err != nil {
		log.Printf("Error de DB. Limpieza de archivos omitida: %v", err)
		if strings.Contains(err.Error(), "unique_violation") {
			return nil, "", "", fmt.Errorf("ya existe un minero registrado con esta cédula o correo electrónico")
		}
		return nil, "", "", fmt.Errorf("fallo al guardar el minero en la base de datos: %w", err)
	}

	return miner, code, qrURL, nil
}

// Paginación
func (s *minerService) GetAllMiners(page, limit int) (*utils.Pagination, error) {
	return s.repo.FindAllPaginated(page, limit)
}

// Validación de archivos
func (s *minerService) validateMinerFiles(minerType models.MinerType, files map[string]*multipart.FileHeader) error {
	baseFields := map[string]models.DocumentField{
		"id_photo_front": {
			FileHeader:       files["id_photo_front"],
			Required:         true,
			MaxSizeBytes:     models.PhotoMaxSize,
			AllowedMimeTypes: []string{"image/jpeg", "image/png"},
		},
		"id_photo_back": {
			FileHeader:       files["id_photo_back"],
			Required:         true,
			MaxSizeBytes:     models.PhotoMaxSize,
			AllowedMimeTypes: []string{"image/jpeg", "image/png"},
		},
		"facial_photo": {
			FileHeader:       files["facial_photo"],
			Required:         true,
			MaxSizeBytes:     models.PhotoMaxSize,
			AllowedMimeTypes: []string{"image/jpeg", "image/png"},
		},
	}

	for fieldName, field := range baseFields {
		if err := utils.ValidateFile(field); err != nil {
			return fmt.Errorf("error de validación en '%s': %w", fieldName, err)
		}
	}

	pdfMimes := []string{"application/pdf", "pdf"}
	var specificFields map[string]models.DocumentField

	switch minerType {
	case models.SubsistenceMiner:
		specificFields = map[string]models.DocumentField{
			"rucon": {
				FileHeader:       files["rucon"],
				Required:         true,
				MaxSizeBytes:     models.RuconMaxSize,
				AllowedMimeTypes: pdfMimes,
			},
			"other_doc": {
				FileHeader:       files["other_doc"],
				Required:         false,
				MaxSizeBytes:     models.SubsistenceOtherMaxSize,
				AllowedMimeTypes: pdfMimes,
			},
		}
	case models.TitularMiner:
		specificFields = map[string]models.DocumentField{
			"exploitation_contract": {
				FileHeader:       files["exploitation_contract"],
				Required:         true,
				MaxSizeBytes:     models.ExploitationContractMaxSize,
				AllowedMimeTypes: pdfMimes,
			},
			"environmental_tool": {
				FileHeader:       files["environmental_tool"],
				Required:         true,
				MaxSizeBytes:     models.EnvironmentalToolMaxSize,
				AllowedMimeTypes: pdfMimes,
			},
			"technical_tool": {
				FileHeader:       files["technical_tool"],
				Required:         true,
				MaxSizeBytes:     models.TechnicalToolMaxSize,
				AllowedMimeTypes: pdfMimes,
			},
		}
	default:
		return fmt.Errorf("tipo de minero no válido")
	}

	for fieldName, field := range specificFields {
		if err := utils.ValidateFile(field); err != nil {
			return fmt.Errorf("error en archivo '%s': %w", fieldName, err)
		}
	}

	return nil
}

// Generar TOTP
func (s *minerService) GenerateTOTP(minerID uuid.UUID) (string, error) {
	miner, err := s.repo.FindByID(minerID)
	if err != nil {
		return "", err
	}
	if miner.TOTPSecret == "" {
		return "", errors.New("el minero no tiene configurado un secreto TOTP")
	}

	code, err := totp.GenerateCodeCustom(miner.TOTPSecret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", err
	}
	return code, nil
}

// Validar TOTP
func (s *minerService) ValidateTOTP(minerID uuid.UUID, code string) (bool, error) {
	miner, err := s.repo.FindByID(minerID)
	if err != nil {
		return false, err
	}
	if miner.TOTPSecret == "" {
		return false, errors.New("el minero no tiene configurado un secreto TOTP")
	}

	valid, err := totp.ValidateCustom(code, miner.TOTPSecret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return false, fmt.Errorf("error al validar código TOTP: %w", err)
	}
	if !valid {
		return false, errors.New("código inválido o expirado")
	}
	return true, nil
}
