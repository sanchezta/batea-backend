package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/batea-fintech/batea-ms-backend/internal/config"
	"github.com/batea-fintech/batea-ms-backend/internal/controller"
	"github.com/batea-fintech/batea-ms-backend/internal/db"
	"github.com/batea-fintech/batea-ms-backend/internal/repository"
	"github.com/batea-fintech/batea-ms-backend/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Cargar Configuración
	cfg := config.LoadConfig()

	// 2. Inicializar Base de Datos (PostgreSQL)
	gormDB, err := db.InitPostgres(cfg)
	if err != nil {
		log.Fatalf("No se pudo inicializar la base de datos: %v", err)
	}
	
	// 3. Inyección de Dependencias (Arquitectura Limpia/Service-Repository)
	minerRepo := repository.NewMinerRepository(gormDB)
	minerService := service.NewMinerService(minerRepo, cfg)
	minerController := controller.NewMinerController(minerService)

	// 4. Configurar Gin
	router := gin.Default()

	// 5. Definir Rutas (Endpoints RESTful)
	v1 := router.Group("/api/v1")
	{
		// Endpoint para registrar un nuevo minero (POST multipart/form-data)
		v1.POST("/miners", minerController.RegisterMiner)
		
		// Endpoint para obtener un minero por ID
		v1.GET("/miners/:id", minerController.GetMinerByID)
	}

	// 6. Manejar el cierre elegante
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatalf("El servidor falló al iniciar: %v", err)
		}
	}()
	log.Printf("Servidor Go/Gin corriendo en el puerto %s...", cfg.Port)

	// Esperar señal de cierre
	<-quit
	log.Println("Apagando el servidor...")
	// Aquí se podría agregar lógica de limpieza de DB o recursos
	log.Println("Servidor apagado.")
}
