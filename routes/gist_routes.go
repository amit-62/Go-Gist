package routes

import (
	"github.com/amit/Go-Gist/controllers"
	"github.com/gin-gonic/gin"
)

type GistRouteController struct {
	gistController controllers.GistController
}

func NewGistRouteController(gistController controllers.GistController) GistRouteController {
	return GistRouteController{gistController: gistController}
}

func (gc *GistRouteController) GistRoute(rg *gin.RouterGroup) {
	router := rg.Group("gists")
	router.GET("/:gistId", gc.gistController.GetGistById)
	router.GET("/:gistId/comments", gc.gistController.GetGistComments)
	router.GET("/:gistId/stargazers", gc.gistController.GetGistStargazers)
}
