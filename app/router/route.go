package router

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"harjonan.id/user-service/app/config"
	"harjonan.id/user-service/app/middleware"
)

func Init(init *config.Initialization) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/", func(c *gin.Context) {
		log.Print("healthy")
		c.JSON(http.StatusOK, gin.H{
			"error":   false,
			"message": "healthy",
			"version": "1.0.0",
		})
	})

	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.APIKeyAuthMiddleware())

	auth := router.Group("/auth")
	{
		auth.POST("/register", init.AuthCtrl.Register)
		auth.POST("/login", init.AuthCtrl.Login)
		auth.POST("/refresh", init.AuthCtrl.Refresh)
		auth.POST("/logout", middleware.JWTAuthMiddleware(), init.AuthCtrl.Logout)
		auth.GET("/me", middleware.JWTAuthMiddleware(), init.AuthCtrl.Me)
	}

	admin := router.Group("/admin", middleware.JWTAuthMiddleware())
	{
		admin.GET("/", init.AdminCtrl.List)
		admin.GET("/:uuid", init.AdminCtrl.Detail)
		admin.POST("/", init.AdminCtrl.Upsert)
		admin.DELETE("/:uuid", init.AdminCtrl.Delete)
	}

	company := router.Group("/company")
	{
		company.GET("/", init.CompanyCtrl.List)
		company.GET("/:uuid", init.CompanyCtrl.Detail)
		company.POST("/", init.CompanyCtrl.Upsert)
		company.DELETE("/:uuid", init.CompanyCtrl.Delete)
	}

	client := router.Group("/client", middleware.JWTAuthMiddleware())
	{
		client.GET("/", init.ClientCtrl.List)
		client.GET("/:uuid", init.ClientCtrl.Detail)
		client.POST("/", init.ClientCtrl.Upsert)
		client.DELETE("/:uuid", init.ClientCtrl.Delete)
	}

	parentMenu := router.Group("/parent-menu", middleware.JWTAuthMiddleware())
	{
		parentMenu.GET("/", init.ParentMenuCtrl.List)
		parentMenu.GET("/:uuid", init.ParentMenuCtrl.Detail)
		parentMenu.POST("/", init.ParentMenuCtrl.Upsert)
		parentMenu.DELETE("/:uuid", init.ParentMenuCtrl.Delete)
	}

	menu := router.Group("/menu", middleware.JWTAuthMiddleware())
	{
		menu.GET("/", init.MenuCtrl.List)
		menu.GET("/:uuid", init.MenuCtrl.Detail)
		menu.POST("/", init.MenuCtrl.Upsert)
		menu.DELETE("/:uuid", init.MenuCtrl.Delete)
	}

	accessMenu := router.Group("/role-menu-access", middleware.JWTAuthMiddleware())
	{
		accessMenu.GET("/", init.RoleAccessMenuCtrl.List)
		accessMenu.GET("/:uuid", init.RoleAccessMenuCtrl.Detail)
		accessMenu.POST("/", init.RoleAccessMenuCtrl.Upsert)
		accessMenu.DELETE("/:uuid", init.RoleAccessMenuCtrl.Delete)
	}

	role := router.Group("/role")
	{
		role.GET("/", init.RoleCtrl.List)
		role.GET("/:uuid", init.RoleCtrl.Detail)
		role.POST("/", init.RoleCtrl.Upsert)
		role.DELETE("/:uuid", init.RoleCtrl.Delete)
	}

	return router
}
