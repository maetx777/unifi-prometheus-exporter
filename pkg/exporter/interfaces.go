package exporter

import "time"

type IDaemonParams interface {
	ControllerLogin() string
	ControllerPassword() string
	ControllerAddress() string
	SnmpExporterAddress() string
	AccessPointsUpdateInterval() time.Duration
	PollTimeout() time.Duration
	ListenPort() int
	Parallel() int
}
