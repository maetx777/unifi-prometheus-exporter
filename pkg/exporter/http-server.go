package exporter

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"sync"
)

func (daemon *Daemon) SetupHttpServer() {
	//настраиваем http-сервер
	//используем mux для роутинга запросов
	router := mux.NewRouter()
	//у нас всего одна ручка /metrics с методом GET
	router.HandleFunc(`/metrics`, daemon.HandleMetrics).Methods(http.MethodGet)
	daemon.httpServer = &http.Server{
		Addr:    fmt.Sprintf(`0.0.0.0:%d`, daemon.params.ListenPort()),
		Handler: router, //используем в сервере mux router
		//прокидываем корневой контекст приложения
		BaseContext: func(listener net.Listener) context.Context {
			return daemon.ctx
		},
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			return daemon.ctx
		},
	}
}

func (daemon *Daemon) RunHttpServer(wg *sync.WaitGroup) {
	logrus.Infoln(`Start http server`)
	defer logrus.Infoln(`Stop http server`)
	defer wg.Done()
	if err := daemon.httpServer.ListenAndServe(); err != nil {
		//если ошибка отличается от server closed (присходит при завершении приложения), значит что то пошло не так
		if !errors.Is(err, http.ErrServerClosed) {
			daemon.fatalChan <- fmt.Errorf(`listen error: %s`, err)
		}
	}
}

func (daemon *Daemon) HandleMetrics(writer http.ResponseWriter, request *http.Request) {
	//получаем сохранённый список точек доступа
	accessPoints := daemon.GetAccessPoints()
	//синхронизатор горутин
	wg := sync.WaitGroup{}
	wg.Add(len(accessPoints))
	//создаём канал для сохранения результатов опроса
	results := make(chan *PollResult, len(accessPoints))
	//запускаем все горутины
	for _, ap := range accessPoints {
		go daemon.pollMetrics(request.Context(), &wg, ap, results)
	}
	//ждём завершения всех горутин
	wg.Wait()
	//отправляем результаты опроса
	for i := 0; i < len(results); i++ {
		result := <-results
		writer.Write(result.Data.Bytes())
	}
}
