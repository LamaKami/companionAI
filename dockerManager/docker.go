package dockerManager

import (
	"fmt"
	"github.com/docker/distribution/context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"io"
	"os"
	"time"
)

type ContainerInformation struct {
	Port    string
	ModelId string
	Version string
	Ip      string
}

// Build takes a buildContextPath which is the path where the Dockerfile lies. The tags are for the name, version, etc.
func Build(buildContextPath string, tags []string) error {

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	buildOpts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       tags,
	}

	buildCtx, err := archive.TarWithOptions(buildContextPath, &archive.TarOptions{})
	if err != nil {
		return err
	}

	resp, err := cli.ImageBuild(ctx, buildCtx, buildOpts)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func Start(imageName string, sourceMountPath string, targetMountPath string, containerPort string) (string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"5000/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: containerPort,
				},
			},
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: sourceMountPath,
				Target: targetMountPath,
			},
		},
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			"5000/tcp": struct{}{},
		},
	}, hostConfig, nil, nil, "")
	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	fmt.Println(resp.ID)
	return resp.ID, nil
}

func StopAll(containerTracker map[string]ContainerInformation) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	for id := range containerTracker {
		fmt.Print("Stopping container ", id, "... ")
		var d time.Duration = -1
		err = cli.ContainerStop(ctx, id, &d)
		if err != nil {
			return err
		}
		fmt.Println("Success")
	}

	//TODO remove container id from global list
	return nil
}

func Stop(containerId string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	fmt.Print("Stopping container ", containerId, "... ")
	var d time.Duration = -1
	err = cli.ContainerStop(ctx, containerId, &d)
	if err != nil {
		return err
	}
	fmt.Println("Success")

	//TODO remove container id from global list
	return nil
}

func GetContainerIp(containerId string) (string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}

	containerInformation, err := cli.ContainerInspect(ctx, containerId)
	if err != nil {
		return "", err
	}

	return containerInformation.NetworkSettings.IPAddress, nil

}
