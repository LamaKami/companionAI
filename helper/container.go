package helper

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

func ContainerAlreadyRunning(modelId string, version string) bool {

	for _, value := range GetContainerTracker() {
		if value.Version == version && value.ModelId == modelId {
			return true
		}
	}
	return false
}

func GetNextPort() string {
	containerTracker := GetContainerTracker()
	log.Println("Searching Port")
	rand.Seed(time.Now().UnixNano())
	min := 49152
	max := 65535
	for {
		port := fmt.Sprintf("%d", rand.Intn(max-min+1)+min)
		log.Print("Checking for port: ", port)

		for _, container := range containerTracker {
			if container.Port == port {
				continue
			}
		}

		conn, err := net.Listen("tcp", ":"+port)
		// TODO
		defer conn.Close()
		if err != nil {
			continue
		}

		log.Println("Found Port")
		return port
	}
}
