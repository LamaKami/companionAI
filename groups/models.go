package groups

import (
	"companionAI/dockerManager"
	"companionAI/helper"
	"companionAI/utils"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
)

// GetModels godoc
// @Tags models
// @Description creates a new folder with the necessary template files
// @Accept json
// @Produce json
// @Success 200 {object} helper.ModelNames
// @Router /models [get]
func GetModels(c *gin.Context) {
	dir, _ := os.Getwd()
	files, err := ioutil.ReadDir(dir + "/mnt/models")
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	names := make([]string, 0)
	for _, file := range files {
		if file.IsDir() {
			names = append(names, file.Name())
		}
	}
	c.JSON(http.StatusOK, helper.ModelNames{
		Names: names,
	})
}

// GetModelTypes godoc
// @Tags models
// @Description returns all available types for creating a new model
// @Accept json
// @Produce json
// @Success 200 {object} helper.ModelTypes
// @Router /models/types [get]
func GetModelTypes(c *gin.Context) {
	types, err := utils.GetModelTypes()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, types)
}

func StopAllContainer(c *gin.Context) {
	containerTracker := helper.GetContainerTracker()
	err := dockerManager.StopAll(containerTracker)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	}
	containerTracker = make(map[string]dockerManager.ContainerInformation)
	c.JSON(http.StatusOK, "Stopped all containers")
}

func GetRunningContainers(c *gin.Context) {
	containerTracker := helper.GetContainerTracker()
	c.JSON(http.StatusOK, containerTracker)
}
