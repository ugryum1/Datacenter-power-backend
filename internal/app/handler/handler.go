package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"datacenter-power/internal/app/config"
	"datacenter-power/internal/app/redis"
	"datacenter-power/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
	Redis      *redis.Client
	Config     *config.Config
}

func NewHandler(r *repository.Repository, rc *redis.Client, c *config.Config) *Handler {
	return &Handler{
		Repository: r,
		Redis:      rc,
		Config:     c,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")

	guest := api.Group("/")
	guest.Use(h.WithOptionalAuth)
	guest.GET("/datacenter-equipment", h.GetDatacenterDevices)
	guest.GET("/datacenter-equipment/reel", h.GetDatacenterDeviceReel)
	guest.GET("/datacenter-equipment/reel/:id", h.GetDatacenterDeviceReelByID)

	api.POST("/datacenter-engineers/sign-up", h.SignUpDatacenterEngineer)
	api.POST("/datacenter-engineers/sign-in", h.SignInDatacenterEngineer)

	creator := api.Group("/")
	creator.Use(h.WithAuthCheck)
	creator.GET("/datacenter-equipment/draft", h.GetDatacenterDeviceDraft)
	creator.POST("/datacenter-equipment", h.CreateDatacenterDevice)
	creator.PUT("/datacenter-equipment/:id/publish", h.PublishDatacenterDevice)
	creator.DELETE("/datacenter-equipment/:id", h.DeleteDatacenterDevice)
	creator.POST("/datacenter-equipment/:id/like", h.LikeDatacenterDevice)
	creator.POST("/datacenter-engineers/sign-out", h.SignOutDatacenterEngineer)
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.AbortWithStatus(errorStatusCode)
}
