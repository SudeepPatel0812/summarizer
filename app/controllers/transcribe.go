package controllers

import (
	"net/http"
	"summarizer/app/client"
	services "summarizer/app/services"
	utils "summarizer/app/utils"

	"github.com/gin-gonic/gin"
)

func DownloadVideoController(c *gin.Context) {
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

	audioExtractorResponse := services.AudioExtractor(videoDownloaderResponse.Data.(string))

	if !audioExtractorResponse.Status {
		c.JSON(http.StatusBadRequest, utils.Response{
			Code:    400,
			Message: audioExtractorResponse.Message,
			Data:    audioExtractorResponse.Data,
			Status:  false,
		})
		return
	}

	transcribeResponse := client.Transcribe(audioExtractorResponse.Data.(map[string]interface{})["audio_path"].(string))
	if !transcribeResponse.Status {
		c.JSON(http.StatusBadRequest, utils.Response{
			Code:    400,
			Message: transcribeResponse.Message,
			Data:    transcribeResponse.Data,
			Status:  false,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Code:    200,
		Message: transcribeResponse.Message,
		Data:    transcribeResponse.Data,
		Status:  transcribeResponse.Status,
	})
}
