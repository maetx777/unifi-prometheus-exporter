package exporter

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func (daemon *Daemon) catchSignals() {
	logrus.Infoln(`Start signals catcher`)
	defer logrus.Infof(`Stop signals catcher`)
	defer daemon.httpServer.Shutdown(daemon.ctx)
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-daemon.ctx.Done():
			return
		case s := <-sigChan:
			logrus.Infof(`Catch %v signal`, s)
			daemon.cancel()
			return
		}
	}
}
