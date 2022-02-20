package helper

type ContainerInformation struct {
	Port    string
	ModelId string
	Version string
	Ip      string
}

var containerTracker = make(map[string]ContainerInformation)

func GetContainerTracker() map[string]ContainerInformation {
	return containerTracker
}

func ResetContainerTracker() {
	containerTracker = make(map[string]ContainerInformation)
}
