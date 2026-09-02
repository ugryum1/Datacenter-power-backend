package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"datacenter-power/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

func (h *Handler) GetTile(ctx *gin.Context) {
	powerFilter := ctx.Query("power")
	minPower := 0
	if powerFilter != "" {
		if v, err := strconv.Atoi(powerFilter); err == nil {
			minPower = v
		}
	}

	devices, err := h.Repository.GetDevices(minPower)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "tile.html", gin.H{
		"active":    "tile",
		"devices":   devices,
		"power":     powerFilter,
		"navFeedID": h.Repository.GetFirstPublishedID(),
	})
}

func (h *Handler) GetFeed(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusBadRequest, "feed.html", gin.H{
			"active":    "feed",
			"error":     "некорректный идентификатор устройства",
			"navFeedID": h.Repository.GetFirstPublishedID(),
		})
		return
	}

	var device repository.Device
	if ctx.Query("next") == "true" {
		device, err = h.Repository.GetNextDevice(id)
	} else {
		device, err = h.Repository.GetDevice(id)
	}
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusNotFound, "feed.html", gin.H{
			"active":    "feed",
			"error":     "устройство не найдено",
			"navFeedID": h.Repository.GetFirstPublishedID(),
		})
		return
	}

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"active":    "feed",
		"device":    device,
		"likes":     device.LikesCount(),
		"navFeedID": h.Repository.GetFirstPublishedID(),
	})
}

func (h *Handler) GetAdd(ctx *gin.Context) {
	draft, err := h.Repository.GetDraft()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "add.html", gin.H{
		"active":    "add",
		"device":    draft,
		"navFeedID": h.Repository.GetFirstPublishedID(),
	})
}
