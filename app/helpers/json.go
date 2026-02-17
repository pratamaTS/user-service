package helpers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/domain/dto"
)

func JsonOK[T any](ctx *gin.Context, msg string, data T) {
	ctx.JSON(http.StatusOK, dto.APIResponse[T]{
		Error:   false,
		Message: msg,
		Data:    data,
	})
}

func JsonErr[T any](ctx *gin.Context, msg string, code int, err error) {
	message := msg
	if err != nil {
		message = msg + ": " + err.Error()
	}

	ctx.JSON(code, dto.APIResponse[T]{
		Error:   true,
		Message: message,
	})
}
