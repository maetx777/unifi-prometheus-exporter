package exporter

import (
	"context"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

type Daemon struct {
	mtx          *sync.RWMutex      //мьютекс, используется для работы со списком точек доступа
	params       IDaemonParams      //параметры cli
	ctx          context.Context    //корневой контекст
	cancel       context.CancelFunc //функция завершения корневого контекста
	accessPoints []*AccessPoint     //список точек доступа
	fatalChan    chan error         //канал с фаталами
	semaphore    chan struct{}      //семафор для ограничения количества горутин
	httpServer   *http.Server       //http-сервер
}

func NewDaemon(params IDaemonParams) *Daemon {
	//создаём инстанс структуры, сразу сохраняем в него параметры и мьютекс
	daemon := &Daemon{params: params, mtx: &sync.RWMutex{}}
	//создаём корневой контекст
	daemon.ctx, daemon.cancel = context.WithCancel(context.Background())
	//канал для фаталов, используется при авторизации
	daemon.fatalChan = make(chan error)
	//семафор для ограничения количества потоков одновременного получения данных от точек доступа
	daemon.semaphore = make(chan struct{}, params.Parallel())
	//настраиваем http-сервер
	daemon.SetupHttpServer()
	return daemon
}

func (daemon *Daemon) Run() error {
	logrus.Infoln(`Daemon start`)
	//при запуске демона сразу запускаем горутину для отлова сигналов завершения
	go daemon.catchSignals()
	//а также прослушку канала с фаталами
	go daemon.fatalsCatcher()
	//переменная для синхронизации горутин
	wg := sync.WaitGroup{}
	wg.Add(2)
	//запускаем фетчер данных от точек доступа
	go daemon.AccessPointsUpdater(&wg)
	//запускаем http-сервер
	go daemon.RunHttpServer(&wg)
	//ждём завршения горутин
	wg.Wait()
	logrus.Infoln(`Daemon stop`)
	return nil
}

func (daemon *Daemon) GetAccessPoints() []*AccessPoint {
	daemon.mtx.RLock()
	defer daemon.mtx.RUnlock()
	return daemon.accessPoints
}

func (daemon *Daemon) SetAccessPoints(items []*AccessPoint) {
	daemon.mtx.Lock()
	defer daemon.mtx.Unlock()
	if len(items) > 0 {
		logrus.Infoln(`Update access points list`)
	}
	for _, item := range items {
		logrus.Infof(`Access point name %s, ip %s`, item.Name, item.Ip)
	}
	daemon.accessPoints = items
}
