package main

import (
	"companionAI/dockerManager"
	"companionAI/docs"
	"companionAI/helper"
	"companionAI/utils"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	cp "github.com/otiai10/copy"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// key is the containerId
var containerTracker = make(map[string]dockerManager.ContainerInformation)

// @BasePath /api/v1

// PredictData godoc
// @Summary
// @Schemes
// @Description generates prediction for the datapoint
// @Param        modelId   path      string  true  "unique id for models"
// @Param        modelVersion   path      string  true  "use version v1 if no other versions exist"
// @Accept json
// @Produce json
// @Success 200 {string} prediction
// @Router /model/predict/{modelId}/{modelVersion} [post]
func PredictData(c *gin.Context) {
	// with the model id we can target different functions therefore each model type must be unique at the start
	containerId := c.Param("containerId")
	containerInformation, contains := containerTracker[containerId]
	if !contains {
		c.JSON(http.StatusBadRequest, "ContainerId does not exist")
		return
	}

	url := fmt.Sprintf("http://%s:5000/predict", containerInformation.Ip)

	HandleRequest(c, "POST", url, c.Request.Body)
}

func TrainModel(c *gin.Context) {
	// TODO handling a training response -> continuous data stream
	containerId := c.Param("containerId")
	containerInformation, contains := containerTracker[containerId]
	if !contains {
		c.JSON(http.StatusBadRequest, "ContainerId does not exist")
		return
	}

	url := fmt.Sprintf("http://%s:5000/train", containerInformation.Ip)

	HandleRequest(c, "GET", url, c.Request.Body)
}

func LoadModel(c *gin.Context) {
	//TODO create function for this:
	containerId := c.Param("containerId")
	containerInformation, contains := containerTracker[containerId]
	if !contains {
		c.JSON(http.StatusBadRequest, "ContainerId does not exist")
		return
	}

	url := fmt.Sprintf("http://%s:5000/load/v1", containerInformation.Ip)

	HandleRequest(c, "GET", url, c.Request.Body)
}

func HandleRequest(c *gin.Context, requestMethod string, url string, payload io.Reader) {
	client := &http.Client{}
	req, err := http.NewRequest(requestMethod, url, payload)

	if err != nil {
		c.JSON(http.StatusBadRequest, fmt.Errorf("error while creating the request %w", err))
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, fmt.Errorf("error while sending the request %w", err))
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, fmt.Errorf("error while reading response body %w", err))
		return
	}
	c.JSON(http.StatusOK, string(body))
}

func CreateNewModel(c *gin.Context) {
	types, err := utils.GetModelTypes()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	var newModel helper.NewModel
	decoder := json.NewDecoder(c.Request.Body)
	err = decoder.Decode(&newModel)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	validModelType := utils.ModelTypeInTypes(types, newModel)

	if !validModelType {
		c.JSON(http.StatusBadRequest, "This model type is currently not supported.")
		return
	}

	dir, _ := os.Getwd()
	files, err := ioutil.ReadDir(dir + "/mnt/models")
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if !utils.CheckStringAlphabet(newModel.Name) {
		c.JSON(http.StatusBadRequest, "model can only use characters from a-z A-Z 0-9")
		return
	}

	for _, file := range files {
		if file.Name() == newModel.Name {
			c.JSON(http.StatusBadRequest, "model with this name already exists")
			return
		}
	}

	err = cp.Copy("mnt/templates/"+newModel.Type, "mnt/models/"+newModel.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, "model was created")
}

func GetModels(c *gin.Context) {
	dir, _ := os.Getwd()
	files, err := ioutil.ReadDir(dir + "/mnt/models")
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	names := make([]string, 0)
	for _, file := range files {
		names = append(names, file.Name())
	}
	c.JSON(http.StatusOK, names)
}

func GetModelTypes(c *gin.Context) {
	types, err := utils.GetModelTypes()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, types)
}

func RemoveModel(c *gin.Context) {
	version := c.Query("version")
	modelId := c.Param("modelId")

	if version != "" {
		fmt.Println(version)
	}
	workingDir, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	err = os.RemoveAll(workingDir + "/mnt/models/" + modelId)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, "model was removed")
}

func AddDataPoints(c *gin.Context) {
	modelId := c.Param("modelId")

	// TODO extract function for the model type config.json
	var dataPoints helper.EntityDataPoints
	decoder := json.NewDecoder(c.Request.Body)
	err := decoder.Decode(&dataPoints)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	for i, value := range dataPoints.EntityDataPoints {
		dataPoints.EntityDataPoints[i].Id = fmt.Sprintf("%x", md5.Sum([]byte(value.Sentence)))
	}

	dir, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	}

	// TODO take correct trainings-data name from config.yml file in data
	dataPath := dir + "/mnt/models/" + modelId + "/data/trainingsData.json"

	var savedData helper.EntityDataPoints
	if err := utils.Load(dataPath, &savedData); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	dataPoints.EntityDataPoints = append(dataPoints.EntityDataPoints, savedData.EntityDataPoints...)

	err = utils.Save(dataPath, dataPoints)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, "Data was saved")
}

type IdBody struct {
	Ids []string `json:"ids"`
}

func DeleteDataPoints(c *gin.Context) {
	modelId := c.Param("modelId")
	var ids IdBody
	decoder := json.NewDecoder(c.Request.Body)
	err := decoder.Decode(&ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	dir, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	}

	// TODO take correct trainings-data name from config.yml file in data
	dataPath := dir + "/mnt/models/" + modelId + "/data/trainingsData.json"

	var savedData helper.EntityDataPoints
	if err := utils.Load(dataPath, &savedData); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	savedData.EntityDataPoints = utils.RemoveElementsFromSlice(savedData.EntityDataPoints, ids.Ids)

	err = utils.Save(dataPath, savedData)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, ids.Ids)

}

func GetDataPoints(c *gin.Context) {
	modelId := c.Param("modelId")

	dir, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	}

	// TODO take correct trainings-data name from config.yml file in data
	dataPath := dir + "/mnt/models/" + modelId + "/data/trainingsData.json"

	var savedData helper.EntityDataPoints
	if err := utils.Load(dataPath, &savedData); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, savedData)
}

func StartContainer(c *gin.Context) {
	version := c.Param("modelVersion")
	modelId := c.Param("modelId")

	if helper.ContainerAlreadyRunning(modelId, version, containerTracker) {
		c.JSON(http.StatusOK, "Container with this modelId and version is already running.")
		return
	}

	dir, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	}

	err = dockerManager.Build(dir+"/mnt/models/"+modelId, []string{modelId})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	port := helper.GetNextPort(containerTracker)

	id, err := dockerManager.Start(modelId, os.Args[1]+"/models/"+modelId, "/mnt", port)

	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	ip, err := dockerManager.GetContainerIp(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, fmt.Errorf("error while trying to get containerIp %w", err))
		return
	}

	containerTracker[id] = dockerManager.ContainerInformation{Port: port, ModelId: modelId, Version: version, Ip: ip}

	c.JSON(http.StatusOK, gin.H{"port": port, "id": id, "ip": ip})
}

func EndContainer(c *gin.Context) {
	containerId := c.Param("containerId")
	err := dockerManager.Stop(containerId)
	if err != nil {
		c.JSON(http.StatusBadRequest, fmt.Errorf("could not stop container %w", err))
		return
	}

	delete(containerTracker, containerId)

	c.JSON(http.StatusOK, "Successfully stopped container!")
}

func GetLabels(c *gin.Context) {
	modelId := c.Param("modelId")
	dir, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	config, err := utils.LoadConfig(dir, modelId)

	c.JSON(http.StatusOK, gin.H{"labels": config.Labels})
}

type LabelBody struct {
	Labels []string `json:"labels"`
}

func AddLabels(c *gin.Context) {
	modelId := c.Param("modelId")
	dir, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	var labels LabelBody
	decoder := json.NewDecoder(c.Request.Body)
	err = decoder.Decode(&labels)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	newLabels, err := utils.AddLabels(dir, modelId, labels.Labels)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"labels": newLabels})
}

func RemoveLabels(c *gin.Context) {
	modelId := c.Param("modelId")
	dir, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	var labels LabelBody
	decoder := json.NewDecoder(c.Request.Body)
	err = decoder.Decode(&labels)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	newLabels, err := utils.RemoveLabels(dir, modelId, labels.Labels)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"labels": newLabels})
}

func StopAllContainer(c *gin.Context) {
	err := dockerManager.StopAll(containerTracker)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	}
	containerTracker = make(map[string]dockerManager.ContainerInformation)
	c.JSON(http.StatusOK, "Stopped all containers")
}

func GetRunningContainers(c *gin.Context) {
	c.JSON(http.StatusOK, containerTracker)
}

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
			modelGroup.POST("/predict/:containerId", PredictData)
			modelGroup.GET("/train/:containerId", TrainModel)
			modelGroup.GET("/load/:containerId", LoadModel)
			modelGroup.POST("/create", CreateNewModel)
			modelGroup.DELETE("/:modelId", RemoveModel)
			modelGroup.POST("/:modelId/:modelVersion/start", StartContainer)
			modelGroup.PUT("/:containerId/stop", EndContainer)
			modelGroup.GET("/:modelId/labels", GetLabels)
			modelGroup.POST("/:modelId/labels", AddLabels)
			modelGroup.DELETE("/:modelId/labels", RemoveLabels)

		}

		modelsGroup := v1.Group("/models")
		{
			modelsGroup.GET("", GetModels)
			modelsGroup.GET("/types", GetModelTypes)
			modelsGroup.PUT("/stopAll", StopAllContainer)
			modelsGroup.GET("/runningContainers", GetRunningContainers)
		}

		dataGroup := v1.Group("/data")
		{
			dataGroup.POST("/:modelId", AddDataPoints)
			dataGroup.GET("/:modelId", GetDataPoints)
			dataGroup.DELETE("/:modelId", DeleteDataPoints)
		}
	}

	server.GET("swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	server.Run()
}
