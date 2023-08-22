package utils

import (
	"net"
	"os"
)

// FreePort finds a free port.
func FreePort() (int, error) {
	address, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}

	defer func() {
		err = listener.Close()
	}()

	return listener.Addr().(*net.TCPAddr).Port, err
}

// DirectoryExist checks if the given directory path exists or not.
func DirectoryExist(path string) bool {
	if stat, err := os.Stat(path); os.IsNotExist(err) || !stat.IsDir() {
		return false
	}
	return true
}
