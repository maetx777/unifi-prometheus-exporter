package exporter

import (
	"bytes"
	"net/http"
	"time"
)

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

type IUnifiClient interface {
	Authorize() (*http.Response, error)
	GetAccessPointsList() (*http.Response, []*AccessPoint, error)
}

type ISnmpExporterClient interface {
	GetMetrics(ap *AccessPoint) (*http.Response, *bytes.Buffer, error)
}
