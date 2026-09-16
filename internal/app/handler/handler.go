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
	router.GET("/dc-equipment", h.GetDatacenterDevices)
	router.GET("/dc-equipment/reel/:id", h.GetDatacenterDevice)
	router.GET("/dc-equipment/draft", h.GetDatacenterDeviceDraft)
	router.POST("/dc-equipment/draft", h.CreateDatacenterDeviceDraft)
	router.POST("/dc-equipment/draft/publish", h.PublishDatacenterDeviceDraft)
	router.POST("/dc-equipment/delete", h.DeleteDatacenterDevice)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/styles", "./resources/styles")
	router.Static("/img", "./resources/img")
	router.Static("/video", "./resources/video")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
