package server

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем или генерируем request-id
		requestID := c.GetHeader("x-request-id")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Сохраняем в контекст Gin
		c.Set("x-request-id", requestID)

		// Добавляем в заголовки ответа
		c.Writer.Header().Set("x-request-id", requestID)

		c.Next()
	}
}
