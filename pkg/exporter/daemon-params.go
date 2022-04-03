package exporter

import (
	"github.com/urfave/cli/v2"
	"time"
)

type DaemonParams struct {
	context *cli.Context
}

func NewDaemonParams(context *cli.Context) *DaemonParams {
	return &DaemonParams{context: context}
}

func (params *DaemonParams) ControllerLogin() string {
	return params.context.String(`controller-login`)
}

func (params *DaemonParams) ControllerPassword() string {
	return params.context.String(`controller-password`)
}

func (params *DaemonParams) ControllerAddress() string {
	return params.context.String(`controller-address`)
}

func (params *DaemonParams) SnmpExporterAddress() string {
	return params.context.String(`snmp-exporter-address`)
}

func (params *DaemonParams) AccessPointsUpdateInterval() time.Duration {
	return params.context.Duration(`access-points-update-interval`)
}

func (params *DaemonParams) PollTimeout() time.Duration {
	return params.context.Duration(`poll-timeout`)
}

func (params *DaemonParams) ListenPort() int {
	return params.context.Int(`listen-port`)
}

func (params *DaemonParams) Parallel() int {
	return params.context.Int(`parallel`)
}
