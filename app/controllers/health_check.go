package controllers

import (
	"summarizer/app/models"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(200, models.Response{
		Code:    200,
		Message: "Service is running",
		Data:    nil,
		Status:  true,
	})
}
