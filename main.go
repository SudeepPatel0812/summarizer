package main

import (
	"summarizer/app/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize the Gin router and set up the routes
	router := gin.Default()
	routes.InitRoutes(router)
	router.Run(":8080")
}
