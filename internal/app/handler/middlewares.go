package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"

	"datacenter-power/internal/app/ds"
)

const jwtPrefix = "Bearer "

const engineerIDContextKey = "engineer_id"

func (h *Handler) WithAuthCheck(ctx *gin.Context) {
	claims, statusCode, err := h.parseDatacenterEngineerJWT(ctx)
	if err != nil {
		h.errorHandler(ctx, statusCode, err)
		return
	}

	ctx.Set(engineerIDContextKey, claims.EngineerID)
}

func (h *Handler) WithOptionalAuth(ctx *gin.Context) {
	claims, _, err := h.parseDatacenterEngineerJWT(ctx)
	if err != nil {
		return
	}

	ctx.Set(engineerIDContextKey, claims.EngineerID)
}

func CurrentEngineerID(ctx *gin.Context) uint {
	return ctx.GetUint(engineerIDContextKey)
}

func (h *Handler) parseDatacenterEngineerJWT(
	ctx *gin.Context,
) (*ds.DatacenterEngineerJWTClaims, int, error) {
	jwtStr := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) {
		return nil, http.StatusForbidden, errors.New("нет токена в заголовке Authorization")
	}
	jwtStr = jwtStr[len(jwtPrefix):]

	err := h.Redis.CheckJWTInBlacklist(ctx.Request.Context(), jwtStr)
	if err == nil {
		return nil, http.StatusForbidden, errors.New("токен в блеклисте")
	}
	if !errors.Is(err, redis.Nil) {
		return nil, http.StatusInternalServerError, err
	}

	token, err := jwt.ParseWithClaims(
		jwtStr,
		&ds.DatacenterEngineerJWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(h.Config.JWT.Token), nil
		},
	)
	if err != nil {
		return nil, http.StatusForbidden, err
	}

	return token.Claims.(*ds.DatacenterEngineerJWTClaims), http.StatusOK, nil
}
