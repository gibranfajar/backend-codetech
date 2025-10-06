package main

import (
	"time"

	"github.com/gibranfajar/backend-codetech/config"
	"github.com/gibranfajar/backend-codetech/controller"
	"github.com/gibranfajar/backend-codetech/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	// koneksi ke database
	config.ConnectDB()

	// validator
	config.InitValidator()

	// inisialisasi router
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://codetech.crx.my.id"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "API CONNECTED SUCCESSFULLY✅",
		})
	})

	// routers
	router.POST("/api/login", controller.Login)
	router.POST("/api/create-user", controller.CreateUser)

	user := router.Group("/api")
	user.GET("/pages", controller.GetAllPages)
	user.GET("/abouts", controller.GetAllAbout)
	user.GET("/services", controller.GetAllServices)
	user.GET("/portfolios", controller.GetAllPortfolio)
	user.GET("/products", controller.GetAllProduct)
	user.GET("/contacts", controller.GetAllContact)
	user.GET("/users", controller.GetUserNotAdmin)
	user.GET("/category-articles", controller.GetAllCategoryArticle)
	user.GET("/articles", controller.GetAllArticle)
	user.GET("/category-faqs", controller.GetAllCategoryFaq)
	user.GET("/faqs", controller.GetAllFaq)
	// update counter views artikel
	user.GET("/articles/:slug/views", controller.IncrementArticleViews)

	// router untuk admin
	protected := router.Group("/api/admin")
	protected.Use(middlewares.AuthMiddleware())
	{
		// route pages
		protected.GET("/pages", controller.GetAllPages)
		protected.POST("/pages", controller.CreatePage)
		protected.PUT("/pages/:id", controller.UpdatePage)
		protected.DELETE("/pages/:id", controller.DeletePage)

		// route about
		protected.GET("/abouts", controller.GetAllAbout)
		protected.POST("/abouts", controller.CreateAbout)
		protected.PUT("/abouts/:id", controller.UpdateAbout)
		protected.DELETE("/abouts/:id", controller.DeleteAbout)

		// route services
		protected.GET("/services", controller.GetAllServices)
		protected.POST("/services", controller.CreateService)
		protected.PUT("/services/:id", controller.UpdateService)
		protected.DELETE("/services/:id", controller.DeleteService)

		// route portfolios
		protected.GET("/portfolios", controller.GetAllPortfolio)
		protected.POST("/portfolios", controller.CreatePortfolio)
		protected.PUT("/portfolios/:id", controller.UpdatePortfolio)
		protected.DELETE("/portfolios/:id", controller.DeletePortfolio)

		// route products
		protected.GET("/products", controller.GetAllProduct)
		protected.POST("/products", controller.CreateProduct)
		protected.PUT("/products/:id", controller.UpdateProduct)
		protected.DELETE("/products/:id", controller.DeleteProduct)

		// route contacts
		protected.GET("/contacts", controller.GetAllContact)
		protected.POST("/contacts", controller.CreateContact)
		protected.PUT("/contacts/:id", controller.UpdateContact)
		protected.DELETE("/contacts/:id", controller.DeleteContact)

		// route users
		protected.GET("/users", controller.GetAllUser)
		protected.POST("/users", controller.CreateUser)
		protected.PUT("/users/:id", controller.UpdateUser)
		protected.DELETE("/users/:id", controller.DeleteUser)
		// get user by is login
		protected.GET("/users/me", controller.GetUser)

		// route category faq
		protected.GET("/category-faqs", controller.GetAllCategoryFaq)
		protected.POST("/category-faqs", controller.CreateCategoryFaq)
		protected.PUT("/category-faqs/:id", controller.UpdateCategoryFaq)
		protected.DELETE("/category-faqs/:id", controller.DeleteCategoryFaq)

		// route faq
		protected.GET("/faqs", controller.GetAllFaq)
		protected.POST("/faqs", controller.CreateFaq)
		protected.PUT("/faqs/:id", controller.UpdateFaq)
		protected.DELETE("/faqs/:id", controller.DeleteFaq)

		// route category articles
		protected.GET("/category-articles", controller.GetAllCategoryArticle)
		protected.POST("/category-articles", controller.CreateCategoryArticle)
		protected.PUT("/category-articles/:id", controller.UpdateCategoryArticle)
		protected.DELETE("/category-articles/:id", controller.DeleteCategoryArticle)

		// route articles
		protected.GET("/articles", controller.GetAllArticle)
		protected.POST("/articles", controller.CreateArticle)
		protected.PUT("/articles/:id", controller.UpdateArticle)
		protected.DELETE("/articles/:id", controller.DeleteArticle)
	}

	// route static untuk menampilkan gambar
	router.Static("/uploads", "uploads")

	router.Run(":8080")

}
