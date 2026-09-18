package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"datacenter-power/internal/app/ds"
)

const (
	currentEngineerID = 1
	powerFilterStepW  = 50
)

func (h *Handler) GetDatacenterDevices(ctx *gin.Context) {
	var devices []ds.DatacenterDevice
	var err error

	powerFilterW, _ := strconv.Atoi(ctx.Query("power"))
	if powerFilterW <= 0 {
		powerFilterW = 0
		devices, err = h.Repository.GetPublishedDatacenterDevices()
	} else {
		devices, err = h.Repository.SearchPublishedDatacenterDevicesByPower(powerFilterW)
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	deviceIDs := make([]uint, 0, len(devices))
	for _, device := range devices {
		deviceIDs = append(deviceIDs, device.ID)
	}
	likesCounts, err := h.Repository.GetDatacenterDevicesLikesCounts(deviceIDs)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "equipment-grid.html", gin.H{
		"activeTab":     "grid",
		"equipmentList": devices,
		"likesCounts":   likesCounts,
		"powerFilterW":  powerFilterW,
		"powerLimitW":   h.powerSliderLimitW(),
		"powerStepW":    powerFilterStepW,
		"reelStartID":   h.Repository.GetFirstPublishedDatacenterDeviceID(),
	})
}

func (h *Handler) GetDatacenterDevice(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logrus.Error(err)
		h.renderReelError(
			ctx,
			http.StatusBadRequest,
			"некорректный идентификатор устройства дата-центра",
		)
		return
	}

	var device ds.DatacenterDevice
	if ctx.Query("next") == "true" {
		device, err = h.Repository.GetNextPublishedDatacenterDevice(id)
	} else {
		device, err = h.Repository.GetPublishedDatacenterDeviceByID(id)
	}
	if err != nil {
		logrus.Error(err)
		h.renderReelError(ctx, http.StatusNotFound, "устройство дата-центра не найдено")
		return
	}

	likesCount, err := h.Repository.GetDatacenterDeviceLikesCount(device.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "equipment-reel.html", gin.H{
		"activeTab":   "reel",
		"equipment":   device,
		"likesCount":  likesCount,
		"reelStartID": h.Repository.GetFirstPublishedDatacenterDeviceID(),
	})
}

func (h *Handler) GetDatacenterDeviceDraft(ctx *gin.Context) {
	draftDevice, err := h.Repository.GetDatacenterDeviceDraft(currentEngineerID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "equipment-draft.html", gin.H{
		"activeTab":   "draft",
		"equipment":   draftDevice,
		"reelStartID": h.Repository.GetFirstPublishedDatacenterDeviceID(),
	})
}

func (h *Handler) CreateDatacenterDeviceDraft(ctx *gin.Context) {
	draftDevice, err := h.Repository.GetDatacenterDeviceDraft(currentEngineerID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if draftDevice == nil {
		_, err = h.Repository.CreateDatacenterDeviceDraft(currentEngineerID, ctx.PostForm("name"))
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	ctx.Redirect(http.StatusFound, "/datacenter-equipment/draft")
}

func (h *Handler) PublishDatacenterDeviceDraft(ctx *gin.Context) {
	powerConsumptionW, err := strconv.Atoi(ctx.PostForm("power"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	heatDissipationBTU, err := strconv.Atoi(ctx.PostForm("heat"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	deviceID, err := h.Repository.PublishDatacenterDeviceDraft(
		currentEngineerID,
		ctx.PostForm("name"),
		ctx.PostForm("short_description"),
		powerConsumptionW,
		heatDissipationBTU,
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, fmt.Sprintf("/datacenter-equipment/reel/%d", deviceID))
}

func (h *Handler) DeleteDatacenterDevice(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.PostForm("device_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.DeleteDatacenterDevice(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	redirectURL := "/datacenter-equipment"
	if powerFilterW, err := strconv.Atoi(ctx.PostForm("power")); err == nil && powerFilterW > 0 {
		redirectURL = fmt.Sprintf("/datacenter-equipment?power=%d", powerFilterW)
	}
	ctx.Redirect(http.StatusFound, redirectURL)
}

func (h *Handler) renderReelError(ctx *gin.Context, statusCode int, message string) {
	ctx.HTML(statusCode, "equipment-reel.html", gin.H{
		"activeTab":   "reel",
		"error":       message,
		"reelStartID": h.Repository.GetFirstPublishedDatacenterDeviceID(),
	})
}

func (h *Handler) powerSliderLimitW() int {
	maxPowerW := h.Repository.GetMaxPublishedPowerConsumptionW()
	if maxPowerW == 0 {
		return powerFilterStepW
	}
	return ((maxPowerW + powerFilterStepW - 1) / powerFilterStepW) * powerFilterStepW
}
