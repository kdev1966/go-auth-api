// routes/routes.go

package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kdev1966/go-auth-api/controllers"
	"github.com/kdev1966/go-auth-api/middleware"
)

func SetupRoutes() *gin.Engine {
	router := gin.Default()

	// Routes publiques
	public := router.Group("/api")
	{
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)
		public.POST("/refresh", controllers.RefreshToken)
	}

	// Routes protégées avec JWT
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/me", controllers.GetMe)                         // accès au profil via l'ID du token
		protected.GET("/users/:id", controllers.GetUserByID)            // admin ou user concerné
		protected.PUT("/users/:id", controllers.UpdateUser)             // admin ou user concerné
		protected.DELETE("/users/:id", controllers.DeleteUser)          // admin ou user concerné
		protected.DELETE("/users/:id/hard", controllers.HardDeleteUser) // admin uniquement
		protected.POST("/users/avatar", controllers.UploadAvatar)       // upload avatar
		protected.GET("/logs", controllers.GetActivityLogs)

		// Routes protégées par IsAdmin uniquement
		admin := protected.Group("")
		admin.Use(middleware.IsAdmin())
		{
			admin.GET("/users", controllers.GetAllUsers)
			admin.PATCH("/users/:id/restore", controllers.RestoreUser)
		}
	}

	return router
}
