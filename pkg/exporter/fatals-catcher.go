package exporter

import (
	"github.com/sirupsen/logrus"
)

func (daemon *Daemon) fatalsCatcher() {
	logrus.Infoln(`Start fatals catcher`)
	defer logrus.Infof(`Stop fatals catcher`)
	for {
		select {
		case err := <-daemon.fatalChan:
			logrus.Fatalln(err)
			return
		}
	}
}
