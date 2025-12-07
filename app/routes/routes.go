package routes

import (
	controllers "summarizer/app/controllers"

	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine) {
	// GET
	r.GET("/health", controllers.HealthCheck)

	// POST
	r.POST("/summarizer", controllers.Summarizer)
}
