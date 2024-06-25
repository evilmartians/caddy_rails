package process

import (
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpstreamProcess(t *testing.T) {
	t.Run("return exit code on exit", func(t *testing.T) {
		p, err := NewUpstreamProcess("false", []string{}, true, "")
		assert.NoError(t, err)

		err = p.Run()
		assert.NoError(t, err)
	})

	t.Run("signal a process to stop it", func(t *testing.T) {
		var err error

		p, err := NewUpstreamProcess("sleep", []string{"10"}, true, "")
		assert.NoError(t, err)

		go func() {
			err = p.Run()
			assert.NoError(t, err)
		}()

		time.Sleep(100 * time.Millisecond)

		err = p.signal(syscall.SIGTERM)
		assert.NoError(t, err)
	})
}
