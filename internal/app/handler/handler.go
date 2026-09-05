package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"datacenter-power/internal/app/repository"
)

const powerFilterStepW = 50

type DatacenterDeviceHandler struct {
	Repository *repository.DatacenterDeviceRepository
}

func NewDatacenterDeviceHandler(r *repository.DatacenterDeviceRepository) *DatacenterDeviceHandler {
	return &DatacenterDeviceHandler{Repository: r}
}

func (h *DatacenterDeviceHandler) ShowEquipmentGrid(ctx *gin.Context) {
	powerFilterW := 0
	if raw := ctx.Query("power"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			powerFilterW = v
		}
	}

	equipmentList, err := h.Repository.ListPublishedDevices(powerFilterW)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "equipment-grid.html", gin.H{
		"activeTab":     "grid",
		"equipmentList": equipmentList,
		"powerFilterW":  powerFilterW,
		"powerLimitW":   h.powerSliderLimitW(),
		"powerStepW":    powerFilterStepW,
		"reelStartID":   h.Repository.FirstPublishedDeviceID(),
	})
}

func (h *DatacenterDeviceHandler) ShowEquipmentReel(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusBadRequest, "equipment-reel.html", gin.H{
			"activeTab":   "reel",
			"error":       "некорректный идентификатор устройства дата-центра",
			"reelStartID": h.Repository.FirstPublishedDeviceID(),
		})
		return
	}

	var equipment repository.DatacenterDevice
	if ctx.Query("next") == "true" {
		equipment, err = h.Repository.GetNextPublishedDevice(id)
	} else {
		equipment, err = h.Repository.GetDeviceByID(id)
	}
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusNotFound, "equipment-reel.html", gin.H{
			"activeTab":   "reel",
			"error":       "устройство дата-центра не найдено",
			"reelStartID": h.Repository.FirstPublishedDeviceID(),
		})
		return
	}

	ctx.HTML(http.StatusOK, "equipment-reel.html", gin.H{
		"activeTab":   "reel",
		"equipment":   equipment,
		"likesCount":  equipment.LikesCount(),
		"reelStartID": h.Repository.FirstPublishedDeviceID(),
	})
}

func (h *DatacenterDeviceHandler) ShowEquipmentDraft(ctx *gin.Context) {
	draftEquipment, err := h.Repository.GetDraftDevice()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "equipment-draft.html", gin.H{
		"activeTab":   "draft",
		"equipment":   draftEquipment,
		"reelStartID": h.Repository.FirstPublishedDeviceID(),
	})
}

func (h *DatacenterDeviceHandler) powerSliderLimitW() int {
	maxPower := h.Repository.MaxPowerConsumptionW()
	if maxPower == 0 {
		return powerFilterStepW
	}
	return ((maxPower + powerFilterStepW - 1) / powerFilterStepW) * powerFilterStepW
}
