package controller

import (
	"log"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/batea-fintech/batea-ms-backend/internal/models"
	"github.com/batea-fintech/batea-ms-backend/internal/service"
	"github.com/gin-gonic/gin"
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
    // 1️⃣ Primero parseamos el formulario
    if err := ctx.Request.ParseMultipartForm(500 * models.Megabyte); err != nil {
        ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{
            "error": "Tamaño del formulario o archivo demasiado grande",
            "details": err.Error(),
        })
        return
    }

    // 2️⃣ Ahora sí, hacemos el bind de los campos de texto
    var req models.CreateMinerRequest
    if err := ctx.ShouldBind(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error": "Datos de entrada inválidos",
            "details": err.Error(),
        })
        return
    }

    // 3️⃣ Mapear archivos
    files := make(map[string]*multipart.FileHeader)
    files["id_photo_front"], _ = ctx.FormFile("id_photo_front")
    files["id_photo_back"], _ = ctx.FormFile("id_photo_back")
    files["facial_photo"], _ = ctx.FormFile("facial_photo")
    files["rucon"], _ = ctx.FormFile("rucon")
    files["other_doc"], _ = ctx.FormFile("other_doc")
    files["exploitation_contract"], _ = ctx.FormFile("exploitation_contract")
    files["environmental_tool"], _ = ctx.FormFile("environmental_tool")
    files["technical_tool"], _ = ctx.FormFile("technical_tool")

    // 4️⃣ Lógica de negocio
    miner, err := c.minerService.CreateMiner(&req, files)
    if err != nil {
        log.Printf("Error al crear minero: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "error": "Fallo en el registro del minero",
            "details": err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusCreated, gin.H{
        "message": "Minero registrado exitosamente y archivos procesados.",
        "miner":   miner,
    })
}

// GetMinerByID obtiene un minero por su ID (ejemplo de endpoint RESTful adicional).
// GET /miners/:id
func (c *MinerController) GetMinerByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	_, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID de minero inválido"})
		return
	}

	// Para esta demostración, no se implementó el método FindByID en el servicio,
	// pero aquí iría la llamada al servicio:
	// miner, err := c.minerService.GetMinerByID(uint(id))
	// if err != nil { ... }
	
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Funcionalidad pendiente: obtener minero por ID " + idStr,
	})
}


// GetAllMiners maneja la solicitud GET para listar mineros con paginación.
// GET /api/v1/miners?page=1&limit=10
func (c *MinerController) GetAllMiners(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	// Llamamos al servicio para obtener la lista paginada
	result, err := c.minerService.GetAllMiners(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
