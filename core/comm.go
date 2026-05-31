package core

import (
	"syscall"
)

type FDComm struct {
	Fd int
}

func (f FDComm) Read(b []byte) (n int, err error) {
	return syscall.Read(f.Fd, b)
}


func (f FDComm) Write(b []byte) (n int, err error) {
	return syscall.Write(f.Fd, b)
}

func (f FDComm) Close() error {
	return syscall.Close(f.Fd)
}
