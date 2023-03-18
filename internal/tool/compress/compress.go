package compress

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

var Compressing string

func Compress(src string, dst string) (err error) {
	cmd := exec.Command(Compressing, "a", "-tzip", "-v512m", "-mx5", dst, src)
	cmd.Stderr = os.Stderr
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
