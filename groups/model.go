package groups

import (
	"companionAI/dockerManager"
	"companionAI/helper"
	"companionAI/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	cp "github.com/otiai10/copy"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// CreateNewModel godoc
// @Tags model
// @Summary create new model
// @Description creates a new folder with the necessary template files
// @Param data body helper.NewModel true "body data"
// @Accept json
// @Produce json
// @Success 200 {string} message
// @Router /model/create [post]
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

// PredictData godoc
// @Tags model
// @Summary predict datapoint
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
	containerTracker := helper.GetContainerTracker()
	containerInformation, contains := containerTracker[containerId]
	if !contains {
		c.JSON(http.StatusBadRequest, "ContainerId does not exist")
		return
	}

	url := fmt.Sprintf("http://%s:5000/predict", containerInformation.Ip)

	handleRequest(c, "POST", url, c.Request.Body)
}

// TrainModel godoc
// @Tags model
// @Summary train model
// @Description trains the machine learning model in the container
// @Param        containerId   path      string  true  "unique id for the container"
// @Accept json
// @Produce json
// @Success 200 {string} message
// @Router /model/train/{containerId} [put]
func TrainModel(c *gin.Context) {
	// TODO handling a training response -> continuous data stream
	containerId := c.Param("containerId")
	containerTracker := helper.GetContainerTracker()
	containerInformation, contains := containerTracker[containerId]
	if !contains {
		c.JSON(http.StatusBadRequest, "ContainerId does not exist")
		return
	}

	url := fmt.Sprintf("http://%s:5000/train", containerInformation.Ip)

	handleRequest(c, "GET", url, c.Request.Body)
}

// LoadModel godoc
// @Tags model
// @Summary load model
// @Description loads the machine learning model in the container
// @Param        containerId   path      string  true  "unique id for the container"
// @Accept json
// @Produce json
// @Success 200 {string} message
// @Router /model/load/{containerId} [put]
func LoadModel(c *gin.Context) {
	//TODO create function for this:
	containerId := c.Param("containerId")
	containerTracker := helper.GetContainerTracker()
	containerInformation, contains := containerTracker[containerId]
	if !contains {
		c.JSON(http.StatusBadRequest, "ContainerId does not exist")
		return
	}

	url := fmt.Sprintf("http://%s:5000/load/v1", containerInformation.Ip)

	handleRequest(c, "GET", url, c.Request.Body)
}

// StartContainer godoc
// @Tags model
// @Summary start container with model
// @Description starts a container for a given model
// @Param        modelId   path      string  true  "unique id for models"
// @Param        modelVersion   path      string  true  "version for the machine learning model"
// @Accept json
// @Produce json
// @Success 200 {object} helper.ContainerInfo
// @Router /model/{modelId}/{modelVersion}/start [post]
func StartContainer(c *gin.Context) {
	version := c.Param("modelVersion")
	modelId := c.Param("modelId")
	containerTracker := helper.GetContainerTracker()

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

	c.JSON(http.StatusOK, helper.ContainerInfo{Id: id, Ip: ip, Port: port})
}

// GetLabels godoc
// @Tags model
// @Summary get labels
// @Description gets labels from the config file
// @Param        modelId   path      string  true  "unique id for models"
// @Accept json
// @Produce json
// @Success 200 {object} helper.LabelBody
// @Router /model/{modelId}/labels [get]
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

// AddLabels godoc
// @Tags model
// @Summary add labels
// @Description add labels to the config file
// @Param        modelId   path      string  true  "unique id for models"
// @Param data body helper.LabelBody true "body data"
// @Accept json
// @Produce json
// @Success 200 {string} message
// @Router /model/{modelId}/labels [post]
func AddLabels(c *gin.Context) {
	modelId := c.Param("modelId")
	dir, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	var labels helper.LabelBody
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

// RemoveLabels godoc
// @Tags model
// @Summary delete labels
// @Description deletes labels from the config file
// @Param        modelId   path      string  true  "unique id for models"
// @Param data body helper.LabelBody true "body data"
// @Accept json
// @Produce json
// @Success 200 {string} message
// @Router /model/{modelId}/labels [delete]
func RemoveLabels(c *gin.Context) {
	modelId := c.Param("modelId")
	dir, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	var labels helper.LabelBody
	decoder := json.NewDecoder(c.Request.Body)
	err = decoder.Decode(&labels)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	_, err = utils.RemoveLabels(dir, modelId, labels.Labels)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, "labels were updated")
}

// RemoveModel godoc
// @Tags model
// @Summary remove model
// @Description deletes a model and the trainings-data from the local file system
// @Param        modelId   path      string  true  "unique id for models"
// @Accept json
// @Produce json
// @Success 200 {string} message
// @Router /model/{modelId} [delete]
func RemoveModel(c *gin.Context) {
	_ = c.Query("version")
	modelId := c.Param("modelId")

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

// EndContainer godoc
// @Tags model
// @Summary stop container
// @Description stops a single container with given id
// @Param        containerId   path      string  true  "unique id for a container"
// @Accept json
// @Produce json
// @Success 200 {string} message
// @Router /model/{containerId}/stop [put]
func EndContainer(c *gin.Context) {
	containerId := c.Param("containerId")
	err := dockerManager.Stop(containerId)
	if err != nil {
		c.JSON(http.StatusBadRequest, fmt.Errorf("could not stop container %w", err))
		return
	}

	containerTracker := helper.GetContainerTracker()
	delete(containerTracker, containerId)

	c.JSON(http.StatusOK, "Successfully stopped container!")
}

// ModelInformation godoc
// @Tags model
// @Summary get model information
// @Description gets all necessary information for a model id
// @Param        modelId   path      string  true  "unique id for models"
// @Accept json
// @Produce json
// @Success 200 {object} helper.ModelInformation
// @Router /model/{modelId} [get]
func ModelInformation(c *gin.Context) {
	modelId := c.Param("modelId")
	var modelInfo helper.ModelInformation

	workingDir, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	err = utils.Load(workingDir+"/mnt/models/"+modelId+"/config.json", &modelInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, modelInfo)
}

func handleRequest(c *gin.Context, requestMethod string, url string, payload io.Reader) {
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
