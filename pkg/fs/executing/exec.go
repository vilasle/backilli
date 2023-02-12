package executing

import (
	"fmt"
	"os/exec"
	"syscall"
)

func ExecCommand(cmd *exec.Cmd) (err error) {
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return fmt.Errorf("Exit Status: %d", status.ExitStatus())
			}
		} else {
			return fmt.Errorf("cmd.Wait: %v", err)
		}
	}
	return err
}
