package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"be-anis/config"
	"be-anis/controller"
	"be-anis/middleware"
	"be-anis/repository"
	"be-anis/service"
)

func main() {
	env, err := config.LoadEnv()
	if err != nil {
		log.Fatalf("failed to load env: %v", err)
	}
	if env.GroqAPIKey == "" {
		log.Fatal("failed to init recommend endpoint: GROQ_API_KEY is required")
	}
	if env.DatabaseURL == "" {
		log.Fatal("failed to init recommend endpoint: DATABASE_URL is required")
	}

	clients, err := config.NewSupabaseClients(env)
	if err != nil {
		log.Fatalf("failed to init supabase: %v", err)
	}

	authRepo := repository.NewAuthRepository(clients)
	authService := service.NewAuthService(authRepo)
	authController := controller.NewAuthController(authService)

	mockRepo := repository.NewMockRepository(clients)
	openAIRepo := repository.NewOpenAIRepository(env.OpenAIAPIKey)
	embeddingRepo := repository.NewEmbeddingRepository(env.MSMultilingualURL)
	storageRepo := repository.NewStorageRepository(env.UploadDir, env.AppBaseURL, env.UploadPublicPath)
	mockService := service.NewMockService(mockRepo, openAIRepo, embeddingRepo, storageRepo)
	mockController := controller.NewMockController(mockService, authService)

	recommendRepo := repository.NewRecommendRepository(env.DatabaseURL)
	groqRepo := repository.NewGroqRepository(env.GroqAPIKey)
	recommendEmbeddingRepo := repository.NewEmbeddingRepository(env.MSMultilingualURL)
	recommendService := service.NewRecommendService(recommendRepo, groqRepo, recommendEmbeddingRepo)
	recommendController := controller.NewRecommendController(recommendService)

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Static(env.UploadPublicPath, env.UploadDir)
	if env.UploadPublicPath != "/img" {
		// Keep legacy public path for backward compatibility.
		r.Static("/img", env.UploadDir)
	}
	if env.UploadPublicPath != "/uploads/img" {
		// Support direct access by legacy/manual path.
		r.Static("/uploads/img", env.UploadDir)
	}
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	authController.RegisterRoutes(r)
	mockController.RegisterRoutes(r)
	recommendController.RegisterRoutes(r)

	log.Printf("server listening on :%s", env.Port)
	if err := r.Run(":" + env.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
