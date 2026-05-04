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

	clients, err := config.NewSupabaseClients(env)
	if err != nil {
		log.Fatalf("failed to init supabase: %v", err)
	}

	authRepo := repository.NewAuthRepository(clients)
	authService := service.NewAuthService(authRepo)
	authController := controller.NewAuthController(authService)

	mockRepo := repository.NewMockRepository(clients)
	openAIRepo := repository.NewOpenAIRepository(env.OpenAIAPIKey)
	embeddingRepo := repository.NewEmbeddingRepository(env.EmbeddingServiceURL)
	storageRepo := repository.NewStorageRepository(env.UploadDir, env.AppBaseURL)
	mockService := service.NewMockService(mockRepo, openAIRepo, embeddingRepo, storageRepo)
	mockController := controller.NewMockController(mockService, authService)

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Static("/img", env.UploadDir)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	authController.RegisterRoutes(r)
	mockController.RegisterRoutes(r)

	log.Printf("server listening on :%s", env.Port)
	if err := r.Run(":" + env.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
