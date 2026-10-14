package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"

	"datacenter-power/internal/app/ds"
)

// SignUpDatacenterEngineer godoc
//
//	@Summary	Регистрация инженера
//	@Tags		datacenter-engineers
//	@Accept		json
//	@Produce	json
//	@Param		request	body		ds.DatacenterEngineerSignUpRequest	true	"Логин и пароль"
//	@Success	201		{object}	ds.DatacenterEngineer
//	@Failure	400
//	@Router		/datacenter-engineers/sign-up [post]
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
		Password: generateHashString(request.Password),
	}
	if err = h.Repository.CreateDatacenterEngineer(&engineer); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, engineer)
}

// SignInDatacenterEngineer godoc
//
//	@Summary		Аутентификация инженера
//	@Description	Возвращает JWT для заголовка Authorization: Bearer <access_token>
//	@Tags			datacenter-engineers
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ds.DatacenterEngineerSignInRequest	true	"Логин и пароль"
//	@Success		200		{object}	ds.DatacenterEngineerSignInResponse
//	@Failure		400
//	@Failure		403
//	@Router			/datacenter-engineers/sign-in [post]
func (h *Handler) SignInDatacenterEngineer(ctx *gin.Context) {
	var request ds.DatacenterEngineerSignInRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	engineer, err := h.Repository.GetDatacenterEngineerByLogin(request.Login)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.errorHandler(ctx, http.StatusForbidden, err)
		return
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if engineer.Password != generateHashString(request.Password) {
		h.errorHandler(ctx, http.StatusForbidden, errors.New("неверный пароль инженера"))
		return
	}

	token := jwt.NewWithClaims(h.Config.JWT.SigningMethod, &ds.DatacenterEngineerJWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(h.Config.JWT.ExpiresIn).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "datacenter-power",
		},
		EngineerID: engineer.ID,
	})

	strToken, err := token.SignedString([]byte(h.Config.JWT.Token))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.DatacenterEngineerSignInResponse{
		ExpiresIn:   h.Config.JWT.ExpiresIn,
		AccessToken: strToken,
		TokenType:   "Bearer",
	})
}

// SignOutDatacenterEngineer godoc
//
//	@Summary		Деавторизация инженера
//	@Description	Добавляет текущий JWT в блеклист Redis
//	@Tags			datacenter-engineers
//	@Security		BearerAuth
//	@Success		200
//	@Failure		403
//	@Router			/datacenter-engineers/sign-out [post]
func (h *Handler) SignOutDatacenterEngineer(ctx *gin.Context) {
	jwtStr := ctx.GetHeader("Authorization")[len(jwtPrefix):]

	err := h.Redis.WriteJWTToBlacklist(ctx.Request.Context(), jwtStr, h.Config.JWT.ExpiresIn)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusOK)
}

func generateHashString(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
