package controller

import (
	"log"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sanchezta/batea-backend/internal/models"
	"github.com/sanchezta/batea-backend/internal/service"
)

// MinerController define los métodos del controlador HTTP.
type MinerController struct {
	minerService service.MinerService
}

// NewMinerController crea una nueva instancia del controlador.
func NewMinerController(s service.MinerService) *MinerController {
	return &MinerController{minerService: s}
}

 
// RegisterMiner maneja la solicitud POST para registrar un nuevo minero con archivos.
// POST /miners (multipart/form-data)
func (c *MinerController) RegisterMiner(ctx *gin.Context) {
	// Parsear formulario
	if err := ctx.Request.ParseMultipartForm(500 * models.Megabyte); err != nil {
		ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error":   "Tamaño del formulario o archivo demasiado grande",
			"details": err.Error(),
		})
		return
	}

	// Obtener user_id del formulario
	userIDStr := ctx.PostForm("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "El user_id proporcionado no es válido"})
		return
	}

	// Bind campos del formulario
	var req models.CreateMinerRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos de entrada inválidos",
			"details": err.Error(),
		})
		return
	}

	// Mapear archivos
	files := make(map[string]*multipart.FileHeader)
	files["id_photo_front"], _ = ctx.FormFile("id_photo_front")
	files["id_photo_back"], _ = ctx.FormFile("id_photo_back")
	files["facial_photo"], _ = ctx.FormFile("facial_photo")
	files["rucon"], _ = ctx.FormFile("rucon")
	files["other_doc"], _ = ctx.FormFile("other_doc")
	files["exploitation_contract"], _ = ctx.FormFile("exploitation_contract")
	files["environmental_tool"], _ = ctx.FormFile("environmental_tool")
	files["technical_tool"], _ = ctx.FormFile("technical_tool")

	// Llamar al servicio (pasando userID)
	miner, code, qrURL, err := c.minerService.CreateMiner(userID, &req, files)
	if err != nil {
		log.Printf("Error al crear minero: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Fallo en el registro del minero",
			"details": err.Error(),
		})
		return
	}

	// Respuesta
	ctx.JSON(http.StatusCreated, gin.H{
		"message":     "Minero registrado exitosamente.",
		"miner":       miner,
		"totp_code":   code,
		"qr_code_url": qrURL,
	})
}

// GetMinerByID obtiene un minero por su ID
// GET /miners/:id
func (c *MinerController) GetMinerByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	_, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID de minero inválido"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Funcionalidad pendiente: obtener minero por ID " + idStr,
	})
}

// GetAllMiners lista todos los mineros con paginación
// GET /api/v1/miners?page=1&limit=10
func (c *MinerController) GetAllMiners(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	result, err := c.minerService.GetAllMiners(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// Obtener el código TOTP actual para un minero (para Flutter)
// GET /miners/:id/totp
func (c *MinerController) GetCurrentTOTP(ctx *gin.Context) {
	idStr := ctx.Param("id")
	minerID, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	code, err := c.minerService.GenerateTOTP(minerID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"totp_code": code})
}
