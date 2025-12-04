package main

import (
	"authService/config"
	"authService/internal/routes"
	"authService/internal/controllers"
	"github.com/gin-gonic/gin"
	"fmt"
)

func main(){
	config.LoadConfig()
	r := gin.Default()

	authController := controllers.NewAuthController()

	routes.AuthRoutes(r, authController)

	r.Run(fmt.Sprintf(":%s", config.App.Port))
}