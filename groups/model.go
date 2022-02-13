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

	c.JSON(http.StatusOK, gin.H{"port": port, "id": id, "ip": ip})
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

	newLabels, err := utils.RemoveLabels(dir, modelId, labels.Labels)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"labels": newLabels})
}

// RemoveModel godoc
// @Tags model
// @Description deletes a model and the trainings-data from the local file system
// @Param        modelId   path      string  true  "unique id for models"
// @Accept json
// @Produce json
// @Success 200 {string} message
// @Router /model/{model-id} [delete]
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
