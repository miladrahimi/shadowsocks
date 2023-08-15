package utils

import (
	"encoding/json"
	"github.com/labstack/gommon/random"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

// IP finds public IP address.
func IP() string {
	f := func() (string, error) {
		client := http.Client{Timeout: 5 * time.Second}
		req, err := client.Get("http://ip-api.com/json")
		if err != nil {
			return "", err
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(req.Body)

		body, err := io.ReadAll(req.Body)
		if err != nil {
			return "", err
		}

		var ip struct {
			Query string `json:"query"`
		}
		err = json.Unmarshal(body, &ip)
		if err != nil {
			return "", err
		}

		return ip.Query, nil
	}

	for i := 0; i < 5; i++ {
		if ip, err := f(); err != nil {
			return ip
		}
	}

	return "127.0.0.1"
}

// Token generates a random token
func Token() string {
	return random.String(32)
}

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
