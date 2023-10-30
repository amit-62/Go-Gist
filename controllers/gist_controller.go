package controllers

import (
	"net/http"

	"gihub.com/amit/Go-Gist/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GistController struct {
	DB *gorm.DB
}

func NewGistController(DB *gorm.DB) GistController {
	return GistController{
		DB: DB,
	}
}

func (gc *GistController) GetGistById(ctx *gin.Context) {
	gistId := ctx.Params.ByName("gistId")

	var gist models.Gist
	result := gc.DB.
		Preload("GistContent").
		First(&gist, "id = ?", gistId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "gist does not exist"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"gist": gist}})
}

func (gc *GistController) GetGistComments(ctx *gin.Context) {
	gistId := ctx.Params.ByName("gistId")

	var comments []models.Comment
	result := gc.DB.Find(&comments, "gist_id = ?", gistId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "gist does not exist"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"comments": comments}})
}

