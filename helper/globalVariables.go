package helper

import "companionAI/dockerManager"

var containerTracker = make(map[string]dockerManager.ContainerInformation)

func GetContainerTracker() map[string]dockerManager.ContainerInformation {
	return containerTracker
}
