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
		company.POST("/fetch", init.CompanyCtrl.List)
		company.GET("/:uuid", init.CompanyCtrl.Detail)
		company.POST("/", init.CompanyCtrl.Upsert)
		company.DELETE("/:uuid", init.CompanyCtrl.Delete)
	}

	client := router.Group("/client", middleware.JWTAuthMiddleware())
	{
		client.POST("/fetch", init.ClientCtrl.List)
		client.GET("/:uuid", init.ClientCtrl.Detail)
		client.POST("/", init.ClientCtrl.Upsert)
		client.DELETE("/:uuid", init.ClientCtrl.Delete)
	}

	parentMenu := router.Group("/parent-menu", middleware.JWTAuthMiddleware())
	{
		parentMenu.POST("/fetch", init.ParentMenuCtrl.List)
		parentMenu.GET("/:uuid", init.ParentMenuCtrl.Detail)
		parentMenu.POST("/", init.ParentMenuCtrl.Upsert)
		parentMenu.DELETE("/:uuid", init.ParentMenuCtrl.Delete)
	}

	menu := router.Group("/menu", middleware.JWTAuthMiddleware())
	{
		menu.POST("/fetch", init.MenuCtrl.List)
		menu.GET("/:uuid", init.MenuCtrl.Detail)
		menu.POST("/", init.MenuCtrl.Upsert)
		menu.DELETE("/:uuid", init.MenuCtrl.Delete)
	}

	accessMenu := router.Group("/role-menu-access", middleware.JWTAuthMiddleware())
	{
		accessMenu.POST("/fetch", init.RoleAccessMenuCtrl.List)
		accessMenu.GET("/:uuid", init.RoleAccessMenuCtrl.Detail)
		accessMenu.POST("/", init.RoleAccessMenuCtrl.Upsert)
		accessMenu.DELETE("/:uuid", init.RoleAccessMenuCtrl.Delete)
	}

	role := router.Group("/role")
	{
		role.POST("/fetch", init.RoleCtrl.List)
		role.GET("/:uuid", init.RoleCtrl.Detail)
		role.POST("/", init.RoleCtrl.Upsert)
		role.DELETE("/:uuid", init.RoleCtrl.Delete)
	}

	subscription := router.Group("/subscriptions", middleware.JWTAuthMiddleware())
	{
		subscription.POST("/upsert", init.SubscriptionCtrl.Upsert)
		subscription.GET("/:uuid", init.SubscriptionCtrl.Detail)
		subscription.POST("/list", init.SubscriptionCtrl.List)
		subscription.DELETE("/:uuid", init.SubscriptionCtrl.Delete)
		subscription.POST("/activate", init.SubscriptionCtrl.ActivateFromMaster)
		subscription.POST("/client-active", init.SubscriptionCtrl.GetClientActiveSubscription)
	}

	files := router.Group("/files", middleware.JWTAuthMiddleware())
	{
		files.POST("/upload", init.FileCtrl.Upload)
		files.GET("/*key", init.FileCtrl.Get) // pakai wildcard biar key bisa ada slash
	}

	clientBranch := router.Group("/client-branches", middleware.JWTAuthMiddleware())
	{
		clientBranch.POST("/fetch", init.ClientBranchCtrl.List)
		clientBranch.GET("/:uuid", init.ClientBranchCtrl.Detail)
		clientBranch.POST("/upsert", init.ClientBranchCtrl.Upsert)
		clientBranch.DELETE("/:uuid", init.ClientBranchCtrl.Delete)
	}

	user := router.Group("/users", middleware.JWTAuthMiddleware())
	{
		user.POST("/fetch", init.UserCtrl.List)
		user.GET("/:uuid", init.UserCtrl.Detail)
		user.POST("/upsert", init.UserCtrl.Upsert)
		user.DELETE("/:uuid", init.UserCtrl.Delete)
	}

	clientUser := router.Group("/client-users", middleware.JWTAuthMiddleware())
	{
		clientUser.POST("/fetch", init.ClientUserCtrl.List)
		clientUser.GET("/:uuid", init.ClientUserCtrl.Detail)
		clientUser.POST("/upsert", init.ClientUserCtrl.Upsert)
		clientUser.DELETE("/:uuid", init.ClientUserCtrl.Delete)
	}

	product := router.Group("/products", middleware.JWTAuthMiddleware())
	{
		product.POST("/fetch", init.ProductCtrl.List)
		product.GET("/:uuid", init.ProductCtrl.Detail)
		product.POST("/upsert", init.ProductCtrl.Upsert)
		product.DELETE("/:uuid", init.ProductCtrl.Delete)
		product.POST("/bulk-upload", init.ProductCtrl.BulkUpload)
	}

	stock := router.Group("/stock-transfers", middleware.JWTAuthMiddleware())
	{
		stock.POST("/fetch", init.StockTransferCtrl.List)
		stock.GET("/:uuid", init.StockTransferCtrl.Detail)
		stock.POST("/request", init.StockTransferCtrl.Create)
		stock.POST("/:uuid/warehouse-approve", init.StockTransferCtrl.WarehouseApprove)
		stock.POST("/:uuid/driver-accept", init.StockTransferCtrl.DriverAccept)
		stock.POST("/:uuid/receive", init.StockTransferCtrl.ReceiveDone)
	}

	pos := router.Group("/pos/transactions", middleware.JWTAuthMiddleware())
	{
		pos.POST("/checkout", init.PosTransactionCtrl.Checkout)
		pos.POST("/fetch", init.PosTransactionCtrl.List)
		pos.GET("/:uuid", init.PosTransactionCtrl.Detail)
		pos.POST("/:uuid/void", init.PosTransactionCtrl.Void)

		// optional: barcode scan cepat
		pos.POST("/scan", init.PosTransactionCtrl.ScanByBarcode)
	}

	return router
}
