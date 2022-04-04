package exporter

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
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

type UnifiClient struct {
	ctx        context.Context
	address    string
	login      string
	password   string
	cookies    []*http.Cookie
	httpClient *http.Client
}

func NewUnifiClient(ctx context.Context, address string, login string, password string) IUnifiClient {
	client := &UnifiClient{
		ctx:      ctx,
		address:  address,
		login:    login,
		password: password,
	}
	client.httpClient = &http.Client{} //http-клиент для работы с апишкой контроллера unifi
	client.httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //unifi использует self-signed сертификат
		},
	}
	return client
}

func (client *UnifiClient) Do(request *http.Request) (*http.Response, error) {
	//устанавливаем от авторизации
	for _, cookie := range client.cookies {
		request.AddCookie(cookie)
	}
	request.Header.Set(`Content-Type`, `application/json`)
	return client.httpClient.Do(request)
}

func (client *UnifiClient) Authorize() (*http.Response, error) {
	url := fmt.Sprintf(`%s/api/login`, client.address)
	buf := bytes.NewBuffer(nil)
	_ = json.NewEncoder(buf).Encode(AuthStruct{
		Strict:   true,
		Password: client.password,
		Remember: true,
		Username: client.login,
	})
	//пробуем авторизоваться, если мы не смогли авторизоваться - отдаём ошибку
	request, err := http.NewRequestWithContext(client.ctx, http.MethodPost, url, buf)
	if err != nil {
		return nil, fmt.Errorf(`controller auth request error: %s`, err)
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf(`controller auth response error: %s`, err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`controller auth invalid status code: %d, %s`, response.StatusCode, response.Status)
	} else {
		//сохраняем куки от авторизации, далее будем их использовать для работы с апишкой
		client.cookies = response.Cookies()
		return response, nil
	}
}

func (client *UnifiClient) GetAccessPointsList() (*http.Response, []*AccessPoint, error) {
	url := fmt.Sprintf(`%s/api/s/default/stat/device`, client.address)
	//создаём запрос к апишке
	request, err := http.NewRequestWithContext(client.ctx, http.MethodGet, url, nil)
	//обрабатываем ошибки
	if err != nil {
		return nil, nil, fmt.Errorf(`ap list request error: %s`, err)
	}
	//совершаем запрос на получение статистики, в ней содержится список точек доступа
	response, err := client.Do(request)
	if err != nil {
		return nil, nil, fmt.Errorf(`ap list response error: %s`, err)
	}
	//на всякий случай проверяем код ответа от апишки
	if response.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf(`ap list invalid status code: %d`, response.StatusCode)
	}
	//пробуем прочитать ответ от апишки
	stat := DeviceStat{}
	err = json.NewDecoder(response.Body).Decode(&stat)
	if err != nil {
		return nil, nil, fmt.Errorf(`ap list json decode error: %s`, err)
	}
	return response, stat.Data, nil
}
