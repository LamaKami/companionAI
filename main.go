package main

import (
	"companionAI/docs"
	"companionAI/groups"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @BasePath /api/v1

func main() {
	// Swagger 2.0 Meta Information
	docs.SwaggerInfo.Title = "CompanionAI"
	docs.SwaggerInfo.Description = "CompanionAI - API for accessing docker containers and editing models."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}

	server := gin.Default()

	v1 := server.Group("/api/v1")
	{
		modelGroup := v1.Group("/model")
		{
			modelGroup.POST("/predict/:containerId", groups.PredictData)
			modelGroup.GET("/train/:containerId", groups.TrainModel)
			modelGroup.GET("/load/:containerId", groups.LoadModel)
			modelGroup.POST("/create", groups.CreateNewModel)
			modelGroup.DELETE("/:modelId", groups.RemoveModel)
			modelGroup.POST("/:modelId/:modelVersion/start", groups.StartContainer)
			modelGroup.PUT("/:containerId/stop", groups.EndContainer)
			modelGroup.GET("/:modelId/labels", groups.GetLabels)
			modelGroup.POST("/:modelId/labels", groups.AddLabels)
			modelGroup.DELETE("/:modelId/labels", groups.RemoveLabels)

		}

		modelsGroup := v1.Group("/models")
		{
			modelsGroup.GET("", groups.GetModels)
			modelsGroup.GET("/types", groups.GetModelTypes)
			modelsGroup.PUT("/stopAll", groups.StopAllContainer)
			modelsGroup.GET("/runningContainers", groups.GetRunningContainers)
		}

		dataGroup := v1.Group("/data")
		{
			dataGroup.POST("/:modelId", groups.AddDataPoints)
			dataGroup.GET("/:modelId", groups.GetDataPoints)
			dataGroup.DELETE("/:modelId", groups.DeleteDataPoints)
		}
	}

	server.GET("swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	server.Run()
}
