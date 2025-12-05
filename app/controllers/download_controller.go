package controllers

import (
	"net/http"
	services "summarizer/app/services"
	utils "summarizer/app/utils"

	"github.com/gin-gonic/gin"
)

func DownloadController(c *gin.Context) {
	var req utils.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Code:    400,
			Message: err.Error(),
			Data:    nil,
			Status:  false,
		})
		return
	}

	videoDownloaderResponse := services.VideoDownloader(req.URL)

	if !videoDownloaderResponse.Status {
		c.JSON(http.StatusBadRequest, utils.Response{
			Code:    400,
			Message: videoDownloaderResponse.Message,
			Data:    videoDownloaderResponse.Data,
			Status:  false,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Code:    200,
		Message: videoDownloaderResponse.Message,
		Data:    videoDownloaderResponse.Data,
		Status:  true,
	})
}
