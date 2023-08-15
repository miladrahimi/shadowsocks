package shadowsocks

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

type Shadowsocks struct {
	command     *exec.Cmd
	logger      *zap.Logger
	binaryPaths map[string]string
	configPath  string
}

func (s *Shadowsocks) binaryPath() string {
	if path, found := s.binaryPaths[runtime.GOOS]; found {
		return path
	}
	return s.binaryPaths["linux"]
}

func (s *Shadowsocks) Run(port int) {
	s.command = exec.Command(
		s.binaryPath(),
		"-config", s.configPath,
		"-metrics", fmt.Sprintf("127.0.0.1:%d", port),
		"--replay_history", "10000",
	)
	s.command.Stderr = os.Stderr
	s.command.Stdout = os.Stdout

	s.logger.Debug("starting the shadowsocks service...")
	if err := s.command.Run(); err != nil {
		s.logger.Fatal("cannot start the shadowsocks service", zap.Error(err))
	}
}

func (s *Shadowsocks) Reconfigure() {
	s.logger.Info("reconfiguring the shadowsocks service...")
	if err := s.command.Process.Signal(syscall.SIGHUP); err != nil {
		s.logger.Fatal("cannot reconfigure the shadowsocks service", zap.Error(err))
	}
}

func (s *Shadowsocks) Shutdown() {
	if err := s.command.Process.Kill(); err != nil {
		s.logger.Error("cannot shutdown the shadowsocks service", zap.Error(err))
	} else {
		s.logger.Info("the shadowsocks service closed successfully")
	}
}

func (s *Shadowsocks) Update(keys []Key) error {
	config := map[string][]Key{"keys": keys}
	content, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	if err = os.WriteFile(s.configPath, content, 0755); err != nil {
		return errors.New(fmt.Sprintf("cannot save %s, err: %v", s.configPath, err))
	}
	return nil
}

func New(l *zap.Logger, cp string, bp map[string]string) *Shadowsocks {
	return &Shadowsocks{
		configPath:  cp,
		logger:      l,
		binaryPaths: bp,
	}
}
