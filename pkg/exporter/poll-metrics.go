package exporter

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"sync"
)

func (daemon *Daemon) pollMetrics(ctx context.Context, wg *sync.WaitGroup, ap *AccessPoint, results chan *PollResult) {
	daemon.semaphore <- struct{}{}        //занимаем очередь в канале
	defer func() { <-daemon.semaphore }() //по завершении работы освобождаем слот
	defer wg.Done()
	ctx, cancel := context.WithTimeout(ctx, daemon.params.PollTimeout()) //создаём локальный контекст с таймаутом
	defer cancel()
	client := http.Client{}
	url := fmt.Sprintf(`%s/snmp?target=%s`, daemon.params.SnmpExporterAddress(), ap.Ip)
	//создаём запрос для похода в snmp-экспортер
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logrus.Errorf(`ap %s (%s) request: %s`, ap.Name, ap.Ip, err)
		return
	}
	//отправляем запрос в snmp-экспортер
	response, err := client.Do(request)
	//проверяем ошибки
	if err != nil {
		logrus.Errorf(`ap %s (%s) response: %s`, ap.Name, ap.Ip, err)
		return
	}
	if response.StatusCode != http.StatusOK {
		logrus.Errorf(`ap %s (%s) invalid status code from snmp-exporter: %d`, ap.Name, ap.Ip, http.StatusInternalServerError)
		return
	}
	//теперь нам осталось прочитать результаты и добавить в каждую метрику наш тег
	buf := bytes.NewBuffer(nil)
	replacedBuffer := bytes.NewBuffer(nil)
	buf.ReadFrom(response.Body)
	pattern := regexp.MustCompile(`^([^{]+){([^}]+)}(.+)$`) //решение "на скорую руку"
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		matches := pattern.FindStringSubmatch(scanner.Text())
		if len(matches) < 4 {
			replacedBuffer.WriteString(scanner.Text() + "\n")
		} else {
			apTag := fmt.Sprintf(`ap_name="%s",ap_ip="%s"`, ap.Name, ap.Ip)
			replacedBuffer.WriteString(fmt.Sprintf("%s{%s,%s}%s\n", matches[1], apTag, matches[2], matches[3]))
		}
	}
	results <- &PollResult{
		AP:   ap,
		Data: replacedBuffer,
	}
}
