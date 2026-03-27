//go:build unix

package procutil

import "syscall"

func IsRunning(pid int) (bool, error) {
	if pid <= 0 {
		return false, nil
	}

	err := syscall.Kill(pid, 0)
	switch err {
	case nil, syscall.EPERM:
		return true, nil
	case syscall.ESRCH:
		return false, nil
	default:
		return false, err
	}
}

func Terminate(pid int) error {
	if pid <= 0 {
		return nil
	}

	return syscall.Kill(pid, syscall.SIGTERM)
}
