package exporter

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

type AuthStruct struct {
	Strict   bool   `json:"strict"`   //не понятно что это, но надо отправлять
	Password string `json:"password"` //пароль
	Remember bool   `json:"remember"` //галка из веб-интерфейса "запомнить меня"
	Username string `json:"username"` //пользователь
}

type AccessPoint struct {
	Ip   string `json:"ip"`   //ipv4 точки доступа
	Name string `json:"name"` //имя точки доступа
}

type DeviceStat struct {
	Data []*AccessPoint `json:"data"` //массив с данными точек доступа
}

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
	client := http.Client{} //http-клиент для работы с апишкой контроллера unifi
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //unifi использует self-signed сертификат
		},
	}
	url := fmt.Sprintf(`%s/api/login`, daemon.params.ControllerAddress())
	buf := bytes.NewBuffer(nil)
	_ = json.NewEncoder(buf).Encode(AuthStruct{
		Strict:   true,
		Password: daemon.params.ControllerPassword(),
		Remember: true,
		Username: daemon.params.ControllerLogin(),
	})
	//далее пробуем авторизоваться, если мы не смогли авторизоваться - считаем это фаталом
	//приложение дальше не может работать, так как нам нужен список точек доступа
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		daemon.fatalChan <- fmt.Errorf(`controller auth request error: %s`, err)
		return
	}
	request.Header.Set(`Content-Type`, `application/json`)
	response, err := client.Do(request)
	if err != nil {
		daemon.fatalChan <- fmt.Errorf(`controller auth response error: %s`, err)
		return
	}
	if response.StatusCode != http.StatusOK {
		daemon.fatalChan <- fmt.Errorf(`controller auth invalid status code: %d, %s`, response.StatusCode, response.Status)
		return
	}
	logrus.Infoln(`Http client authorized`)
	//сохраняем куки от авторизации, далее будем их использовать для работы с апишкой
	var cookies = response.Cookies()
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
				url = fmt.Sprintf(`%s/api/s/default/stat/device`, daemon.params.ControllerAddress())
				//создаём запрос к апишке
				request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
				//обрабатываем ошибки
				if err != nil {
					logrus.Errorf(`ap list request error: %s`, err)
					return
				}
				//устанавливаем от авторизации
				for _, cookie := range cookies {
					request.AddCookie(cookie)
				}
				//совершаем запрос на получение статистики, в ней содержится список точек доступа
				response, err := client.Do(request)
				if err != nil {
					logrus.Errorf(`ap list response error: %s`, err)
					return
				}
				//на всякий случай проверяем код ответа от апишки
				if response.StatusCode != http.StatusOK {
					logrus.Errorf(`ap list invalid status code: %d`, response.StatusCode)
					return
				}
				//пробуем прочитать ответ от апишки
				stat := DeviceStat{}
				err = json.NewDecoder(response.Body).Decode(&stat)
				if err != nil {
					logrus.Errorf(`ap list json decode error: %s`, err)
					return
				}
				//сохраняем список точек в структуру
				daemon.SetAccessPoints(stat.Data)
			}()
		}
	}
}
