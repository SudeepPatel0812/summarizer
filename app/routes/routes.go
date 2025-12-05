package routes

import (
	controllers "summarizer/app/controllers"

	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine) {
	r.GET("/health", controllers.HealthCheck)
	r.POST("/download", controllers.DownloadController)
}
