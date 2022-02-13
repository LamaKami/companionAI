package utils

import (
	"bytes"
	"companionAI/helper"
	"encoding/json"
	"io"
	"os"
	"sync"
)

type Config struct {
	Modeltype     string   `json:"model-type"`
	NewestVersion string   `json:"newest-version"`
	Labels        []string `json:"labels"`
}

var Marshal = func(v interface{}) (io.Reader, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

var Unmarshal = func(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

var lock sync.Mutex

func Save(path string, v interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r, err := Marshal(v)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}

func Load(path string, v interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return Unmarshal(f, v)
}

func LoadConfig(dir string, modelId string) (Config, error) {

	configPath := dir + "/mnt/models/" + modelId + "/config.json"

	var config Config
	if err := Load(configPath, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func AddLabels(dir string, modelId string, labels []string) ([]string, error) {

	config, err := LoadConfig(dir, modelId)
	if err != nil {
		return nil, err
	}

	config.Labels = append(config.Labels, labels...)
	configPath := dir + "/mnt/models/" + modelId + "/config.json"

	err = Save(configPath, config)
	if err != nil {
		return nil, err
	}

	return config.Labels, nil
}

func RemoveLabels(dir string, modelId string, labelsToRemove []string) ([]string, error) {

	config, err := LoadConfig(dir, modelId)
	if err != nil {
		return nil, err
	}

	configPath := dir + "/mnt/models/" + modelId + "/config.json"
	var newLabels []string
	for _, ele := range config.Labels {
		found := false
		for _, removeEle := range labelsToRemove {
			if removeEle == ele {
				found = true
				break
			}
		}
		if !found {
			newLabels = append(newLabels, ele)
		}
	}
	config.Labels = newLabels

	err = Save(configPath, config)
	if err != nil {
		return nil, err
	}

	return config.Labels, nil
}

func GetModelTypes() (helper.ModelTypes, error) {
	dir, err := os.Getwd()
	if err != nil {
		return helper.ModelTypes{}, err
	}

	var modelTypes helper.ModelTypes
	err = Load(dir+"/mnt/information/modelTypes.json", &modelTypes)
	if err != nil {
		return helper.ModelTypes{}, err
	}

	return modelTypes, nil
}

func ModelTypeInTypes(types helper.ModelTypes, newModel helper.NewModel) bool {
	for _, modelType := range types.ModelTypes {
		if modelType.Name == newModel.Type {
			return true
		}
	}
	return false
}
