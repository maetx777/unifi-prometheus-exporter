package exporter

import (
	"context"
	"github.com/sirupsen/logrus"
	"sync"
)

/*
AccessPointsUpdater - функция для горутины, которая получает от контроллера unifi список точек доступа и сохраняем в структуре
*/
func (daemon *Daemon) AccessPointsUpdater(wg *sync.WaitGroup) {
	logrus.Infoln(`Start access points updater`)
	defer logrus.Infof(`Stop access points updater`)
	defer wg.Done()

	//задаём контекст для функции на основе глобального контекста
	ctx, cancel := context.WithCancel(daemon.ctx)
	defer cancel()

	client := NewUnifiClient(ctx, daemon.params.ControllerAddress(), daemon.params.ControllerLogin(), daemon.params.ControllerPassword())
	if _, err := client.Authorize(); err != nil {
		daemon.fatalChan <- err
	} else {
		logrus.Infoln(`Http client authorized`)
	}

	//работаем в лупе, имея ввиду что контекст приложения может быть закрыт в любой момент
	for {
		select {
		case <-ctx.Done():
			return
		default:
			func() {
				//после отработки шага засыпаем на заданное количество времени
				//обычный sleep не используем, т.к. он не совместим с контекстом
				defer SleepWithContext(ctx, daemon.params.AccessPointsUpdateInterval())
				if _, stat, err := client.GetAccessPointsList(); err != nil {
					logrus.Errorln(err)
				} else {
					//сохраняем список точек в структуру
					daemon.SetAccessPoints(stat)
				}
			}()
		}
	}
}
