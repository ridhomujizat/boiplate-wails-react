//go:build !windows

package helper

import "os/exec"

func newSystemCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}
