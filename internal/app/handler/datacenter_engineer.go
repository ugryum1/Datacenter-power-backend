package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"datacenter-power/internal/app/ds"
)

const currentEngineerID uint = 1

func CurrentEngineerID() uint {
	return currentEngineerID
}

func (h *Handler) SignUpDatacenterEngineer(ctx *gin.Context) {
	var request ds.DatacenterEngineerSignUpRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	exists, err := h.Repository.DatacenterEngineerLoginExists(request.Login)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if exists {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("логин инженера уже занят"))
		return
	}

	engineer := ds.DatacenterEngineer{
		Login:    request.Login,
		Password: request.Password,
	}
	if err = h.Repository.CreateDatacenterEngineer(&engineer); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, engineer)
}

func (h *Handler) SignInDatacenterEngineer(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}

func (h *Handler) SignOutDatacenterEngineer(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}
