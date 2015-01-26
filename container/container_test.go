package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test OSExec really needs to mock, don't tell anyone I did this.
func TestOSExecSuccess(t *testing.T) {
	output, err := OSExec("echo", "hello,", "world!")
	assert.Nil(t, err)
	assert.Equal(t, output, "hello, world!\n", "Testing a call to echo")
}

func TestOSExecFailure(t *testing.T) {
	_, err := OSExec("echosdfjhsdfh")
	assert.Error(t, err)
}
