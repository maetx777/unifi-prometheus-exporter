package exporter

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"regexp"
)

type SnmpExporterClient struct {
	address    string
	httpClient *http.Client
	ctx        context.Context
}

func NewSnmpExporterClient(ctx context.Context, address string) ISnmpExporterClient {
	client := &SnmpExporterClient{ctx: ctx, address: address}
	client.httpClient = &http.Client{}
	return client
}

func (client *SnmpExporterClient) Do(request *http.Request) (*http.Response, error) {
	return client.httpClient.Do(request)
}

func (client *SnmpExporterClient) GetMetrics(ap *AccessPoint) (*http.Response, *bytes.Buffer, error) {
	url := fmt.Sprintf(`%s/snmp?target=%s`, client.address, ap.Ip)
	//создаём запрос для похода в snmp-экспортер
	request, err := http.NewRequestWithContext(client.ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf(`ap %s (%s) request: %s`, ap.Name, ap.Ip, err)
	}
	//отправляем запрос в snmp-экспортер
	response, err := client.Do(request)
	//проверяем ошибки
	if err != nil {
		return nil, nil, fmt.Errorf(`ap %s (%s) response: %s`, ap.Name, ap.Ip, err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf(`ap %s (%s) invalid status code from snmp-exporter: %d`, ap.Name, ap.Ip, http.StatusInternalServerError)
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
	return response, replacedBuffer, nil
}
