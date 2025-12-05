package utils

type Request struct {
    URL string `json:"url" binding:"required"`
}