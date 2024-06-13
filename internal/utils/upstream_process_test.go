package utils

import (
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpstreamProcess(t *testing.T) {
	t.Run("return exit code on exit", func(t *testing.T) {
		p, _ := NewUpstreamProcess("false", []string{}, true, "")
		exitCode, err := p.Run()

		assert.NoError(t, err)
		assert.Equal(t, 1, exitCode)
	})

	t.Run("signal a process to stop it", func(t *testing.T) {
		var exitCode int
		var err error

		p, _ := NewUpstreamProcess("sleep", []string{"10"}, true, "")

		go func() {
			exitCode, err = p.Run()
		}()

		<-p.Started
		p.signal(syscall.SIGTERM)

		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)
	})
}
