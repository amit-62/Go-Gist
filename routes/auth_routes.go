package routes

import (
	"gihub.com/amit/Go-Gist/controllers"	
	"gihub.com/amit/Go-Gist/middleware"
	"github.com/gin-gonic/gin"
)

type AuthRouteController struct {
	authController controllers.AuthController
}

func NewAuthRouteController(authController controllers.AuthController) AuthRouteController {
	return AuthRouteController{
		authController: authController,
	}
}

func (rc *AuthRouteController) AuthRoute(rg *gin.RouterGroup) {
	router := rg.Group("/auth")

	router.POST("/register", rc.authController.SignUpUser)
	router.POST("/login", rc.authController.SignInUser)
	router.GET("/refresh", rc.authController.RefreshAccessToken)
	router.GET("/logout", middleware.DeserializeUser(), rc.authController.LogoutUser)
	router.GET("/verifyemail/:verificationCode", rc.authController.VerifyEmail)
	router.POST("/forgotpassword", rc.authController.ForgotPassword)
	router.PATCH("/resetpassword/:resetToken", rc.authController.ResetPassword)
}
