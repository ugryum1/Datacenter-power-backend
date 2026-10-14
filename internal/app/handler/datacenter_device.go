package handler

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"datacenter-power/internal/app/ds"
	"datacenter-power/internal/app/repository"
)

var datacenterImageExtensions = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

var datacenterVideoExtensions = map[string]string{
	"video/webm": ".webm",
	"video/mp4":  ".mp4",
}

// GetDatacenterDevices godoc
//
//	@Summary		Список опубликованных устройств
//	@Description	is_creator = 1, если устройство создал авторизованный инженер
//	@Tags			datacenter-equipment
//	@Produce		json
//	@Param			power	query	int	false	"Мощность от, Вт"
//	@Success		200		{array}	ds.DatacenterDeviceCardSerializer
//	@Router			/datacenter-equipment [get]
func (h *Handler) GetDatacenterDevices(ctx *gin.Context) {
	var devices []ds.DatacenterDevice
	var err error

	powerFilterW, _ := strconv.Atoi(ctx.Query("power"))
	if powerFilterW <= 0 {
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

	serializers := make([]ds.DatacenterDeviceCardSerializer, 0, len(devices))
	for _, device := range devices {
		serializers = append(
			serializers,
			ds.NewDatacenterDeviceCardSerializer(
				device,
				likesCounts[device.ID],
				CurrentEngineerID(ctx),
			),
		)
	}

	ctx.JSON(http.StatusOK, serializers)
}

// GetDatacenterDeviceReel godoc
//
//	@Summary	Первое устройство ленты
//	@Tags		datacenter-equipment
//	@Produce	json
//	@Success	200	{object}	ds.DatacenterDeviceCardSerializer
//	@Failure	404
//	@Router		/datacenter-equipment/reel [get]
func (h *Handler) GetDatacenterDeviceReel(ctx *gin.Context) {
	device, err := h.Repository.GetFirstPublishedDatacenterDevice()
	if err != nil {
		h.repositoryErrorHandler(ctx, err)
		return
	}

	h.respondDatacenterDeviceCard(ctx, device)
}

// GetDatacenterDeviceReelByID godoc
//
//	@Summary		Устройство ленты по id
//	@Description	При next=true возвращается следующее устройство после id
//	@Tags			datacenter-equipment
//	@Produce		json
//	@Param			id		path		int		true	"id устройства"
//	@Param			next	query		string	false	"true — следующее устройство"
//	@Success		200		{object}	ds.DatacenterDeviceCardSerializer
//	@Failure		400
//	@Failure		404
//	@Router			/datacenter-equipment/reel/{id} [get]
func (h *Handler) GetDatacenterDeviceReelByID(ctx *gin.Context) {
	var device ds.DatacenterDevice

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	if ctx.Query("next") == "true" {
		device, err = h.Repository.GetNextPublishedDatacenterDevice(id)
	} else {
		device, err = h.Repository.GetPublishedDatacenterDeviceByID(id)
	}
	if err != nil {
		h.repositoryErrorHandler(ctx, err)
		return
	}

	h.respondDatacenterDeviceCard(ctx, device)
}

// GetDatacenterDeviceDraft godoc
//
//	@Summary	Черновик устройства текущего инженера
//	@Tags		datacenter-equipment
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	ds.DatacenterDeviceSerializer
//	@Failure	403
//	@Failure	404
//	@Router		/datacenter-equipment/draft [get]
func (h *Handler) GetDatacenterDeviceDraft(ctx *gin.Context) {
	draftDevice, err := h.Repository.GetDatacenterDeviceDraft(CurrentEngineerID(ctx))
	if err != nil {
		h.repositoryErrorHandler(ctx, err)
		return
	}

	h.respondDatacenterDevice(ctx, http.StatusOK, draftDevice)
}

// CreateDatacenterDevice godoc
//
//	@Summary		Добавление устройства с фото и видео
//	@Description	Создаёт черновик текущего инженера, файлы сохраняются в MinIO
//	@Tags			datacenter-equipment
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			name	formData	string	true	"Название устройства"
//	@Param			image	formData	file	true	"Фото (jpeg, png, webp)"
//	@Param			video	formData	file	true	"Видео (webm, mp4)"
//	@Success		201		{object}	ds.DatacenterDeviceSerializer
//	@Failure		400
//	@Failure		403
//	@Router			/datacenter-equipment [post]
func (h *Handler) CreateDatacenterDevice(ctx *gin.Context) {
	err := ctx.Request.ParseMultipartForm(2 << 20)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	_, err = h.Repository.GetDatacenterDeviceDraft(CurrentEngineerID(ctx))
	if err == nil {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("черновик устройства уже существует"))
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	name := ctx.Request.FormValue("name")
	if name == "" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("не указано название устройства"))
		return
	}

	imageHeader, err := ctx.FormFile("image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	videoHeader, err := ctx.FormFile("video")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	imageContentType, imageExtension, err := detectDatacenterMediaType(
		imageHeader,
		datacenterImageExtensions,
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	videoContentType, videoExtension, err := detectDatacenterMediaType(
		videoHeader,
		datacenterVideoExtensions,
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	device, err := h.Repository.CreateDatacenterDeviceDraft(CurrentEngineerID(ctx), name)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	err = h.Repository.AddDatacenterDeviceMedia(
		&device,
		repository.DatacenterMediaFile{
			Header:      imageHeader,
			ContentType: imageContentType,
			Filename:    fmt.Sprintf("datacenter-device-%d-image%s", device.ID, imageExtension),
		},
		repository.DatacenterMediaFile{
			Header:      videoHeader,
			ContentType: videoContentType,
			Filename:    fmt.Sprintf("datacenter-device-%d-video%s", device.ID, videoExtension),
		},
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.respondDatacenterDevice(ctx, http.StatusCreated, device)
}

// PublishDatacenterDevice godoc
//
//	@Summary	Публикация черновика устройства
//	@Tags		datacenter-equipment
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		id		path		int									true	"id черновика"
//	@Param		request	body		ds.DatacenterDevicePublishRequest	true	"Поля устройства"
//	@Success	200		{object}	ds.DatacenterDeviceSerializer
//	@Failure	400
//	@Failure	403
//	@Failure	404
//	@Router		/datacenter-equipment/{id}/publish [put]
func (h *Handler) PublishDatacenterDevice(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var request ds.DatacenterDevicePublishRequest
	if err = ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.PublishDatacenterDevice(id, CurrentEngineerID(ctx), request)
	if err != nil {
		h.repositoryErrorHandler(ctx, err)
		return
	}

	device, err := h.Repository.GetPublishedDatacenterDeviceByID(id)
	if err != nil {
		h.repositoryErrorHandler(ctx, err)
		return
	}

	h.respondDatacenterDevice(ctx, http.StatusOK, device)
}

// DeleteDatacenterDevice godoc
//
//	@Summary	Логическое удаление устройства текущего инженера
//	@Tags		datacenter-equipment
//	@Security	BearerAuth
//	@Param		id	path	int	true	"id устройства"
//	@Success	200
//	@Failure	400
//	@Failure	403
//	@Failure	404
//	@Router		/datacenter-equipment/{id} [delete]
func (h *Handler) DeleteDatacenterDevice(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.DeleteDatacenterDevice(id, CurrentEngineerID(ctx))
	if err != nil {
		h.repositoryErrorHandler(ctx, err)
		return
	}

	ctx.Status(http.StatusOK)
}

// LikeDatacenterDevice godoc
//
//	@Summary		Лайк устройства от текущего инженера
//	@Description	like = 1 ставит лайк, like = 0 снимает
//	@Tags			datacenter-equipment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int								true	"id устройства"
//	@Param			request	body		ds.DatacenterDeviceLikeRequest	true	"Лайк"
//	@Success		200		{object}	ds.DatacenterDeviceSerializer
//	@Failure		400
//	@Failure		403
//	@Failure		404
//	@Router			/datacenter-equipment/{id}/like [post]
func (h *Handler) LikeDatacenterDevice(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var request ds.DatacenterDeviceLikeRequest
	if err = ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	device, err := h.Repository.GetPublishedDatacenterDeviceByID(id)
	if err != nil {
		h.repositoryErrorHandler(ctx, err)
		return
	}

	if *request.Like == 1 {
		err = h.Repository.LikeDatacenterDevice(CurrentEngineerID(ctx), device.ID)
	} else {
		err = h.Repository.UnlikeDatacenterDevice(CurrentEngineerID(ctx), device.ID)
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.respondDatacenterDevice(ctx, http.StatusOK, device)
}

func (h *Handler) respondDatacenterDevice(
	ctx *gin.Context,
	statusCode int,
	device ds.DatacenterDevice,
) {
	likesCount, err := h.Repository.GetDatacenterDeviceLikesCount(device.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(statusCode, ds.NewDatacenterDeviceSerializer(device, likesCount))
}

func (h *Handler) respondDatacenterDeviceCard(ctx *gin.Context, device ds.DatacenterDevice) {
	likesCount, err := h.Repository.GetDatacenterDeviceLikesCount(device.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(
		http.StatusOK,
		ds.NewDatacenterDeviceCardSerializer(device, likesCount, CurrentEngineerID(ctx)),
	)
}

func (h *Handler) repositoryErrorHandler(ctx *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	h.errorHandler(ctx, http.StatusInternalServerError, err)
}

func detectDatacenterMediaType(
	header *multipart.FileHeader,
	extensions map[string]string,
) (string, string, error) {
	file, err := header.Open()
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return "", "", err
	}

	contentType := http.DetectContentType(buffer[:n])
	extension, ok := extensions[contentType]
	if !ok {
		return "", "", fmt.Errorf("недопустимый тип файла: %s", contentType)
	}
	return contentType, extension, nil
}
