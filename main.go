package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"gopkg.in/yaml.v2"
)

// Config структура для хранения конфигурации серверов
type Config struct {
	Servers []string `yaml:"servers"`
}

// sendCommandToServers функция отправки команды на сервера
func sendCommandToServers(servers []string, command string) {
	var wg sync.WaitGroup
	for _, server := range servers {
		wg.Add(1)
		go func(serverURL string) {
			defer wg.Done()

			// Игнорирование ошибок сертификата TLS
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr}

			url := serverURL + command
			resp, err := client.Get(url)
			if err != nil {
				fmt.Printf("Error sending command to server %s: %v\n", serverURL, err)
				return
			}
			defer resp.Body.Close()
			fmt.Printf("Response from server %s: %s\n", serverURL, resp.Status)

			// Чтение ответа от сервера
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response body from server %s: %v\n", serverURL, err)
				return
			}

			// Отображение ответа от сервера
			fmt.Printf("Response from server %s: %s\n", serverURL, string(body))
		}(server)
	}
	wg.Wait()
}

func main() {
	// Определение аргументов командной строки
	configFile := flag.String("config", "", "path to config file")
	command := flag.String("command", "", "command to send to servers")
	flag.Parse()

	// Проверка наличия аргументов
	if *configFile == "" || *command == "" {
		fmt.Println("Usage: go run main.go -config=config.yaml -command='/add_user?user_name=user_t3'")
		return
	}

	// Чтение конфигурационного файла
	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return
	}

	// Декодирование YAML-конфигурации
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("Error decoding config file: %v\n", err)
		return
	}

	// Отправка команды на серверы
	sendCommandToServers(config.Servers, *command)
}
