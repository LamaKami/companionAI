package groups

import (
	"companionAI/helper"
	"companionAI/utils"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

// AddDataPoints godoc
// @Tags data
// @Description adds multiple data points to the trainings-data
// @Param        modelId   path      string  true  "unique id for models"
// @Param data body helper.EntityDataPoints true "id can be ignored"
// @Accept json
// @Produce json
// @Success 200 {string} message
// @Router /data/{modelId} [get]
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

// DeleteDataPoints godoc
// @Tags data
// @Description removes multiple data points from the trainings-data
// @Param        modelId   path      string  true  "unique id for models"
// @Param data body helper.IdBody true "datapoints which should be removed"
// @Accept json
// @Produce json
// @Success 200 {string} message
// @Router /data/{modelId} [delete]
func DeleteDataPoints(c *gin.Context) {
	modelId := c.Param("modelId")
	var ids helper.IdBody
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

	c.JSON(http.StatusOK, "Deleted")

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
