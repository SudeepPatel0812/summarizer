package controllers

import (
	"net/http"
	"summarizer/app/client"
	"summarizer/app/models"
	"summarizer/app/services"
	"summarizer/app/utils"

	"github.com/gin-gonic/gin"
)

// Summarize the audio context from video.
func Summarizer(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: err.Error(),
			Data:    nil,
			Status:  false,
		})
		return
	}

	videoDownloaderResponse := services.VideoDownloader(req.URL)

	if !videoDownloaderResponse.Status {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: videoDownloaderResponse.Message,
			Data:    videoDownloaderResponse.Data,
			Status:  false,
		})
		return
	}

	audioExtractorResponse := services.AudioExtractor(videoDownloaderResponse.Data.(string))

	if !audioExtractorResponse.Status {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: audioExtractorResponse.Message,
			Data:    audioExtractorResponse.Data,
			Status:  false,
		})
		return
	}

	transcribeResponse := client.Summarizer(audioExtractorResponse.Data.(map[string]interface{})["audio_path"].(string))
	if !transcribeResponse.Status {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: transcribeResponse.Message,
			Data:    transcribeResponse.Data,
			Status:  false,
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: transcribeResponse.Message,
		Data:    transcribeResponse.Data,
		Status:  transcribeResponse.Status,
	})

	defer utils.ClearMediaDirectories()
}
