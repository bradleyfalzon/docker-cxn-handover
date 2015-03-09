package container

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	errDockerIDNotSet              = errors.New("docker container id not set")
	errDockerIDNotFound            = "docker container id not found"
	errDockerIDMultipleFound       = "docker container id matched multiple results"
	errDockerIDNotRunning          = errors.New("docker container id is not running")
	errDockerIDMultiplePortsMapped = "docker container has a public port mapped to multiple private ports"
	errDockerIDNonPublishedPort    = "docker container has configured but unpublished or unexposed port"
	errDockerIDCantParsePorts      = "docker container id has exposed or published ports that cannot be parsed"
)

// Docker struct to implement the Container interface
type Docker struct {
	Container
	containerID string
	osExec      CmdRunner
	statePaused,
	stateRunning bool
}

// Structs for docker inspect
type port struct {
	HostPort string
}
type networkSettings struct {
	IPAddress string
	Ports     map[string][]port
}
type state struct {
	Paused, Running bool
}
type dockerInspect struct {
	NetworkSettings networkSettings
	State           state
}

// NewDocker creates a new docker
func NewDocker(containerID string, cr CmdRunner) (d *Docker) {
	d = &Docker{containerID: containerID, osExec: cr}
	d.IsValid()
	d.Inspect()
	return
}

// IsValid performs basic checks of container ID to ensure ID is valid before being used
func (d *Docker) IsValid() (err error) {
	if d.containerID == "" {
		return errDockerIDNotSet
	}
	return
}

// IsRunning checks to ensure a container is currently running (not stopped/paused)
func (d *Docker) IsRunning() (err error) {
	if d.stateRunning && !d.statePaused {
		return
	}
	return errDockerIDNotRunning
}

// Inspect uses docker inspect command to fetch information about the container
func (d *Docker) Inspect() (err error) {
	stdout, err := d.osExec("docker", "inspect", d.containerID)
	if err != nil {
		return
	}

	var inspect []dockerInspect
	err = json.Unmarshal([]byte(stdout), &inspect)

	if len(inspect) == 0 {
		err = fmt.Errorf("%s: %s", errDockerIDNotFound, d.containerID)
		return
	}

	if len(inspect) > 1 {
		err = fmt.Errorf("%s: %s", errDockerIDMultipleFound, d.containerID)
		return
	}

	d.statePaused = inspect[0].State.Paused
	d.stateRunning = inspect[0].State.Running
	d.ipAddress = inspect[0].NetworkSettings.IPAddress

	for public, private := range inspect[0].NetworkSettings.Ports {
		var ipPort IPPort
		ipPort, err = parseInspectPorts(public, private)
		if err != nil {
			if strings.Contains(err.Error(), errDockerIDNonPublishedPort) {
				// Ignore nonpublished port errors, they're not really a problem
				continue
			}
			return
		}
		d.ipPorts = append(d.ipPorts, ipPort)
	}

	return
}

func parseInspectPorts(public string, private []port) (ipPort IPPort, err error) {

	// Parse the public side of the port mapping eg: 80/tcp
	publicSplit := strings.Split(public, "/")
	if len(publicSplit) == 1 || len(publicSplit) > 2 {
		err = fmt.Errorf("%s: %s", errDockerIDCantParsePorts, public)
		return
	}

	var publicPort uint64
	publicPort, err = strconv.ParseUint(publicSplit[0], 10, 16)
	if err != nil {
		return
	}

	// Parse the private side of the port mapping
	if len(private) == 0 {
		// Container has a port configured in DockerFile, but not published
		err = fmt.Errorf("%s for public port: %s", errDockerIDNonPublishedPort, public)
		return
	}
	if len(private) > 1 {
		err = fmt.Errorf("%s for public port: %s", errDockerIDMultiplePortsMapped, public)
		return
	}

	var privatePort uint64
	privatePort, err = strconv.ParseUint(private[0].HostPort, 10, 16)
	if err != nil {
		return
	}

	ipPort = IPPort{
		PublicPort:  uint16(publicPort),
		PrivatePort: uint16(privatePort),
		Proto:       publicSplit[1],
	}

	return
}
