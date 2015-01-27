package container

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test New Docker

func TestNewDocker(t *testing.T) {
	c := NewDocker("Test", OSExec)

	if c.id != "Test" {
		t.Error("NewDocker failed")
	}
}

// Test Docker IsValid

func TestDockerIsValid(t *testing.T) {
	d := NewDocker("", OSExec)
	err := d.IsValid()
	assert.Equal(t, err, errDockerIDNotSet, "Did not detect blank container ID as invalid")

	d = NewDocker("something", OSExec)
	err = d.IsValid()
	assert.Nil(t, err, "Detected valid container ID as invalid")
}

// Test Docker IsRunning

func TestDockerIsRunning(t *testing.T) {
	d := &Docker{stateRunning: true, statePaused: false}
	err := d.IsRunning()
	assert.Nil(t, err, "Failed to find running Container ID")
}
func TestDockerIsRunningNot(t *testing.T) {
	d := &Docker{stateRunning: true, statePaused: true}
	assert.Equal(t, d.IsRunning(), errDockerIDNotRunning)

	d = &Docker{stateRunning: false, statePaused: true}
	assert.Equal(t, d.IsRunning(), errDockerIDNotRunning)

	d = &Docker{stateRunning: false, statePaused: false}
	assert.Equal(t, d.IsRunning(), errDockerIDNotRunning)
}

// Test Docker GetInternalIP

func DockerInspect(string, ...string) (string, error) {
	inspect := `[{
	    "NetworkSettings": {
			"IPAddress": "172.17.0.3",
			"Ports": {
				"443/tcp": null,
				"80/tcp": [
					{
						"HostIp": "0.0.0.0",
						"HostPort": "49154"
					}
				]
			}
		},
		"State": {
			"Paused": false,
			"Running": true
		}
	}]`

	return inspect, nil
}
func TestDockerInspect(t *testing.T) {
	d := NewDocker("somecontaineridnotrunning", DockerInspect)
	err := d.Inspect()
	assert.Nil(t, err)
	assert.True(t, d.stateRunning, "Container is running")
	assert.False(t, d.statePaused, "Container is paused")
	assert.Equal(t, d.ipAddress, "172.17.0.3", "Network IP is set correctly")

	ipPorts := []IPPort{{PublicPort: 80, PrivatePort: 49154, Proto: "tcp"}}
	assert.Equal(t, d.ipPorts, ipPorts, "IP Ports match correctly")
}

func DockerInspectParseError(string, ...string) (string, error) {
	inspect := `[{
	    "NetworkSettings": {
			"Ports": {
				"": null
			}
		}
	}]`

	return inspect, nil
}
func TestDockerInspectParseError(t *testing.T) {
	d := NewDocker("somecontaineridnotrunning", DockerInspectParseError)
	err := d.Inspect()
	assert.Contains(t, err.Error(), errDockerIDCantParsePorts)
}

func DockerInspectExecError(string, ...string) (string, error) {
	return "", errors.New("Error happened")
}
func TestDockerInspectExecError(t *testing.T) {
	d := NewDocker("somecontainerdoesntwork", DockerInspectExecError)
	err := d.Inspect()
	assert.Error(t, err)
}

func DockerInspectNotFound(string, ...string) (string, error) {
	return "", nil
}
func TestDockerInspectNotFound(t *testing.T) {
	d := NewDocker("somecontainerdoesntwork", DockerInspectNotFound)
	err := d.Inspect()
	assert.Contains(t, err.Error(), errDockerIDNotFound)
}

func DockerInspectMultipleFound(string, ...string) (string, error) {
	return "[{},{}]", nil
}
func TestDockerInspectMultipleFound(t *testing.T) {
	d := NewDocker("somecontainerdoesntwork", DockerInspectMultipleFound)
	err := d.Inspect()
	assert.Contains(t, err.Error(), errDockerIDMultipleFound)
}

// Docker Parse Inspect

func TestDockerParseInspectPortsInvalidPublic(t *testing.T) {
	_, err := parseInspectPorts("", []port{})
	assert.Contains(t, err.Error(), errDockerIDCantParsePorts)

	_, err = parseInspectPorts("40/tcp/udp", []port{})
	assert.Contains(t, err.Error(), errDockerIDCantParsePorts)

	_, err = parseInspectPorts("65537/tcp", []port{})
	assert.Error(t, err)

	_, err = parseInspectPorts("-1/tcp", []port{})
	assert.Error(t, err)
}

func TestDockerParseInspectPortsUnpublished(t *testing.T) {
	_, err := parseInspectPorts("80/tcp", []port{})
	assert.Contains(t, err.Error(), errDockerIDNonPublishedPort)
	assert.Contains(t, err.Error(), "80/tcp")
}

func TestDockerParseInspectPortsMultiMapped(t *testing.T) {
	_, err := parseInspectPorts("81/tcp", []port{{HostPort: "4567"}, {HostPort: "1234"}})
	assert.Contains(t, err.Error(), errDockerIDMultiplePortsMapped)
	assert.Contains(t, err.Error(), "81/tcp")
}

func TestDockerParseInspectPortsInvalidPrivate(t *testing.T) {
	_, err := parseInspectPorts("82/tcp", []port{{HostPort: "65537"}})
	assert.Error(t, err)
	_, err = parseInspectPorts("82/tcp", []port{{HostPort: "-1"}})
	assert.Error(t, err)
}
