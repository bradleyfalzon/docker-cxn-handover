package container

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

var (
	errContainerCmd = "Error running external command"
)

// Container interface that other compatible container software, such
// as Docker, LXC, Rocket should implement
type Container struct {
	ipAddress string
	ipPorts   []IPPort
}

// IPPort mapping for a container's published or expose ports
type IPPort struct {
	// Port number that should be published for external access
	PublicPort uint16
	// Private Port container is listening for requests on
	PrivatePort uint16
	// Protocol used by port, such as tcp
	Proto string
}

// CmdRunner is used to mock OSExec
type CmdRunner func(prog string, args ...string) (string, error)

// OSExec executes a command and return the string output or an error
func OSExec(prog string, args ...string) (string, error) {
	cmd := exec.Command(prog, args...)

	var (
		stdErr, stdOut bytes.Buffer
	)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("%s: %s, args: %s, err: %s, stderr: %s", errContainerCmd, prog, strings.Join(args, " "), err, stdErr.String())
		return stdOut.String(), err
	}

	return stdOut.String(), nil
}
