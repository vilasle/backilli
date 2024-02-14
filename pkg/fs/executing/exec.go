package executing

import (
	"fmt"
	"io"
	"os/exec"
	"syscall"
)

func Execute(command string, 
	out io.Writer, 
	err io.Writer,  
	args ...string) error {
	
	cmd := exec.Command(command, args...)
	return execCommand(cmd)
}

func execCommand(cmd *exec.Cmd) (err error) {
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
