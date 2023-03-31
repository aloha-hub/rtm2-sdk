package rtm2_sdk

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"os/exec"
	"syscall"
	"time"
)

type rtmSidecar struct {
	ctx    context.Context
	cancel context.CancelFunc

	cmd  string
	args []string

	process *exec.Cmd
	errChan chan error

	lg *zap.Logger
}

func (s *rtmSidecar) Start() <-chan error {
	p := exec.Command(s.cmd, s.args...)
	s.process = p
	go s.loop()
	return s.errChan
}

func (s *rtmSidecar) StartV1() <-chan error {
	p := exec.CommandContext(s.ctx, s.cmd, s.args...)
	//p.Cancel = func() error {
	//	if err := p.Process.Signal(syscall.SIGTERM); err != nil {
	//		s.lg.Error("fail to sigterm sidecar, we should kill it", zap.Error(err))
	//	}
	//	return p.Process.Kill()
	//}
	err := p.Start()
	if err != nil {
		close(s.errChan)
	} else {
	}
	s.process = p
	go s.loop()
	return s.errChan
}

func (s *rtmSidecar) Stop() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered:", r)
		}
	}()
	if s.process.Process != nil {
		if err := s.process.Process.Signal(syscall.SIGTERM); err != nil {
			s.lg.Error("fail to sigterm sidecar, we should kill it", zap.Error(err))
			if err = s.process.Process.Kill(); err != nil {
				s.lg.Error("fail to kill process", zap.Error(err))
				time.Sleep(5 * time.Second)
			}
		} else {
			time.Sleep(time.Second)
		}
	}
}

func (s *rtmSidecar) loop() {
	err := s.process.Run()
	if err != nil {
		s.lg.Error("process stopped, it should be a fatal error", zap.Error(err))
		s.cancel()
		s.errChan <- err
	} else {
		close(s.errChan)
	}
}

func createSidecar(ctx context.Context, lg *zap.Logger, execPath string, port int32) *rtmSidecar {
	c, cancel := context.WithCancel(ctx)
	return &rtmSidecar{ctx: c, cancel: cancel, cmd: fmt.Sprintf("%s/rtm2-wrapper.exe", execPath),
		args: []string{fmt.Sprintf("--port=%d", port), "--mode=1"}, errChan: make(chan error, 1), lg: lg}
}
