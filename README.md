Пост на habr.com - https://habr.com/ru/post/658863/

```
NAME:
   exporter - экспортер snmp-метрик от точек доступа unifi

USAGE:
   exporter [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --controller-login value               логин от exporter-контроллера [$CONTROLLER_LOGIN]
   --controller-password value            пароль от exporter-контроллера [$CONTROLLER_PASSWORD]
   --controller-address value             адрес exporter-контроллера (default: "https://127.0.0.1:8443") [$CONTROLLER_ADDRESS]
   --snmp-exporter-address value          адрес snmp-экспортера (default: "http://snmp-exporter:9116") [$SNMP_EXPORTER_ADDRESS]
   --access-points-update-interval value  интервал обновления списка точек (default: 1h0m0s) [$ACCESS_POINTS_UPDATE_INTERVAL]
   --listen-port value                    порт прослушки http-сервера (default: 8080) [$LISTEN_PORT]
   --parallel value                       количество потоков для опроса точек-доступа (default: 10) [$PARALLEL]
   --poll-timeout value                   таймаут для опроса точек доступа (default: 15s) [$POLL_TIMEOUT]
   --help, -h                             show help (default: false)
```