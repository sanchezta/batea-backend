package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sanchezta/batea-backend/internal/config"
	"github.com/sanchezta/batea-backend/internal/controller"
	"github.com/sanchezta/batea-backend/internal/db"
	"github.com/sanchezta/batea-backend/internal/repository"
	"github.com/sanchezta/batea-backend/internal/service"
)

func main() {
	// 1. Cargar Configuración
	cfg := config.LoadConfig()

	// 2. Inicializar Base de Datos (PostgreSQL)
	gormDB, err := db.InitPostgres(cfg)
	if err != nil {
		log.Fatalf("No se pudo inicializar la base de datos: %v", err)
	}

	// 3. Inyección de dependencias (Arquitectura limpia)
	userRepo := repository.NewUserRepository(gormDB)
	minerRepo := repository.NewMinerRepository(gormDB)

	userService := service.NewUserService(userRepo)
	minerService := service.NewMinerService(minerRepo, userRepo, cfg)

	userController := controller.NewUserController(userService, minerService)
	minerController := controller.NewMinerController(minerService)

	// 4. Configurar router de Gin
	router := gin.Default()

	// 5. Definir rutas RESTful
	v1 := router.Group("/api/v1")
	{
		// Registro de usuario
		v1.POST("/users/register", userController.RegisterUser)

		// Rutas de mineros
		v1.POST("/miners", minerController.RegisterMiner)
		v1.GET("/miners/:id", minerController.GetMinerByID)
		v1.GET("/miners", minerController.GetAllMiners)
		v1.GET("/miners/:id/totp", minerController.GetCurrentTOTP)
	}

	// 6. Manejar cierre elegante
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatalf("El servidor falló al iniciar: %v", err)
		}
	}()
	log.Printf("Servidor Go/Gin corriendo en el puerto %s...", cfg.Port)

	<-quit
	log.Println("Apagando el servidor...")
	log.Println("Servidor apagado.")
}
