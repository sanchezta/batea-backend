package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sanchezta/batea-backend/internal/models"
	"github.com/sanchezta/batea-backend/internal/service"
)

type UserController struct {
	userService  service.UserService
	minerService service.MinerService
}

func NewUserController(u service.UserService, m service.MinerService) *UserController {
	return &UserController{u, m}
}

// POST /api/register
func (ctrl *UserController) RegisterUser(c *gin.Context) {
	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ctrl.userService.RegisterUser(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}
