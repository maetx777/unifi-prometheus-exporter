package exporter

import (
	"bytes"
	"context"
	"github.com/sirupsen/logrus"
	"sync"
)

type PollResult struct {
	AP   *AccessPoint
	Data *bytes.Buffer
}

func (daemon *Daemon) pollMetrics(ctx context.Context, wg *sync.WaitGroup, ap *AccessPoint, results chan *PollResult) {
	daemon.semaphore <- struct{}{}        //занимаем очередь в канале
	defer func() { <-daemon.semaphore }() //по завершении работы освобождаем слот
	defer wg.Done()
	ctx, cancel := context.WithTimeout(ctx, daemon.params.PollTimeout()) //создаём локальный контекст с таймаутом
	defer cancel()
	
	client := NewSnmpExporterClient(ctx, daemon.params.SnmpExporterAddress())
	if _, data, err := client.GetMetrics(ap); err != nil {
		logrus.Errorln(err)
	} else {
		results <- &PollResult{
			AP:   ap,
			Data: data,
		}
	}
}
