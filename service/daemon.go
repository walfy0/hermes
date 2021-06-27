package service

import (
	"os"
	"syscall"

	"github.com/sirupsen/logrus"
)

const DaemonFlag = "--daemon"

func StartDaemon() (int, error) {
	isDaemon := false
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == DaemonFlag {
			isDaemon = true
		}
	}
	if isDaemon {
		logrus.Infoln(os.Getppid(), os.Getpid())
		_ = syscall.Umask(0)
		sid, s_errno := syscall.Setsid()
		if s_errno != nil || sid < 0 {
			panic(s_errno)
		}
		// os.Chdir("/")
		return 0, nil
	} else {
		//fork a child process
		args := make([]string, 0, len(os.Args)+1)
		args = append(args, os.Args...)
		args = append(args, DaemonFlag)
		execSpec := &syscall.ProcAttr{
			Env:   os.Environ(),
			Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
		}
		pid, err := syscall.ForkExec(os.Args[0], args, execSpec)
		if err != nil {
			return -1, err
		}
		return pid, nil
	}
}
