package helper

import (
	"companionAI/utils"
	"os"
)

type ModelTypes struct {
	ModelTypes []ModelType `json:"modelTypes"`
}

type ModelType struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type NewModel struct {
	Name string
	Type string
}

func GetModelTypes() (ModelTypes, error) {
	dir, err := os.Getwd()
	if err != nil {
		return ModelTypes{}, err
	}

	var modelTypes ModelTypes
	err = utils.Load(dir+"/mnt/information/modelTypes.json", &modelTypes)
	if err != nil {
		return ModelTypes{}, err
	}

	return modelTypes, nil
}

func ModelTypeInTypes(types ModelTypes, newModel NewModel) bool {
	for _, modelType := range types.ModelTypes {
		if modelType.Name == newModel.Type {
			return true
		}
	}
	return false
}
