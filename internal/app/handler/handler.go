package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"datacenter-power/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	api := router.Group("/api")

	api.GET("/datacenter-equipment", h.GetDatacenterDevices)
	api.GET("/datacenter-equipment/reel", h.GetDatacenterDeviceReel)
	api.GET("/datacenter-equipment/reel/:id", h.GetDatacenterDeviceReel)
	api.GET("/datacenter-equipment/draft", h.GetDatacenterDeviceDraft)
	api.POST("/datacenter-equipment", h.CreateDatacenterDevice)
	api.PUT("/datacenter-equipment/:id/publish", h.PublishDatacenterDevice)
	api.DELETE("/datacenter-equipment/:id", h.DeleteDatacenterDevice)
	api.POST("/datacenter-equipment/:id/like", h.LikeDatacenterDevice)

	api.POST("/datacenter-engineers/sign-up", h.SignUpDatacenterEngineer)
	api.POST("/datacenter-engineers/sign-in", h.SignInDatacenterEngineer)
	api.POST("/datacenter-engineers/sign-out", h.SignOutDatacenterEngineer)
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.AbortWithStatus(errorStatusCode)
}
