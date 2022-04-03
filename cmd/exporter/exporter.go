package main

import (
	"github.com/maetx777/unifi-prometheus-exporter/pkg/exporter"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"time"
)

func main() {
	app := cli.App{
		Name:  `exporter`,
		Usage: `экспортер snmp-метрик от точек доступа unifi`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "controller-login",
				Usage:    `логин от exporter-контроллера`,
				EnvVars:  []string{`CONTROLLER_LOGIN`},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "controller-password",
				Usage:    `пароль от exporter-контроллера`,
				EnvVars:  []string{`CONTROLLER_PASSWORD`},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "controller-address",
				Usage:   `адрес exporter-контроллера`,
				EnvVars: []string{`CONTROLLER_ADDRESS`},
				Value:   `https://127.0.0.1:8443`,
			},
			&cli.StringFlag{
				Name:    "snmp-exporter-address",
				Usage:   `адрес snmp-экспортера`,
				EnvVars: []string{`SNMP_EXPORTER_ADDRESS`},
				Value:   `http://snmp-exporter:9116`,
			},
			&cli.DurationFlag{
				Name:    "access-points-update-interval",
				Usage:   `интервал обновления списка точек`,
				EnvVars: []string{`ACCESS_POINTS_UPDATE_INTERVAL`},
				Value:   time.Duration(1) * time.Hour,
			},
			&cli.IntFlag{
				Name:    "listen-port",
				Usage:   `порт прослушки http-сервера`,
				EnvVars: []string{`LISTEN_PORT`},
				Value:   8080,
			},
			&cli.IntFlag{
				Name:    "parallel",
				Usage:   `количество потоков для опроса точек-доступа`,
				EnvVars: []string{`PARALLEL`},
				Value:   10,
			},
			&cli.DurationFlag{
				Name:    "poll-timeout",
				Usage:   `таймаут для опроса точек доступа`,
				EnvVars: []string{`POLL_TIMEOUT`},
				Value:   time.Duration(15) * time.Second,
			},
		},
		Action: func(context *cli.Context) error {
			return exporter.NewDaemon(exporter.NewDaemonParams(context)).Run()
		},
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatalf(`application error: %s`, err)
	}
}
