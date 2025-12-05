package controllers

import (
	utils "summarizer/app/utils"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(200, utils.Response{
		Code:    200,
		Message: "Service is running",
		Data:    nil,
		Status:  true,
	})
}
