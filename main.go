package main

import (
	"fmt"
	"github.com/nats-io/stan.go"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Название канала ля NATS-STREAMING
var channelName = "WBChannel"

// Идентификаторы кластера и клиента
var clusterID = "WB-cluster"
var clientID = "plan9t"

// URL NATS Streaming сервера
var natsURL = "nats://localhost:4222"

// Подключение к NATS Streaming
var sc stan.Conn
var err error
var MyCache = NewCache()

func init() {
	// Инициализация подключения к NATS Streaming
	var err error
	sc, err = stan.Connect(
		clusterID,
		clientID,
		stan.NatsURL(natsURL),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Успешная инициализация")
}

func main() {

	fmt.Println("PROGRAMM STARTED")

	// Работа с КЭШ

	MyCache.AddOrders(GetOrdersFromPostgreSQL())

	// Запуск HTTP сервера на порту и старт прослушки порта 3333 в горутине
	go startHTTPServer()

	// Настройка обработчика для POST запросов
	http.HandleFunc("/forpost", handlePostRequest)

	// Проверка подключения к серверу NATS-streaming
	if sc.NatsConn().IsConnected() {
		fmt.Println("ПОДКЛЮЧЕНО К СЕРВЕРУ NATS-STREAMING")
	} else {
		fmt.Println("Не подключено к серверу NATSSTREAMING")
	}

	// Создание канала (темы) при публикации сообщения
	err = sc.Publish(channelName, []byte("Канал "+channelName+" был создан"))
	if err != nil {
		log.Fatal(err)
	}

	// Подписка на канал и прослушка канала (здесь мы парсим JSON и записываем данные в PostgreSQL
	subscription, err := sc.Subscribe(channelName, func(msg *stan.Msg) {
		fmt.Printf("Получено сообщение из канала '%s': %s\n", channelName, string(msg.Data))

		// Вызывай функцию для парсинга JSON и сохранения в базу данных
		orderData, err := parseJSON(msg.Data)
		if err != nil {
			fmt.Println("Ошибка при парсинге JSON:", err)
			return
		}

		err = SaveToPostgreSQL(orderData)
		if err != nil {
			fmt.Println("Ошибка при сохранении в PostgreSQL:", err)
			return
		}

		// Запись в кэш
		MyCache.AddOrder(orderData)

	}, stan.DurableName(clientID))

	if err != nil {
		log.Fatal(err)
	}

	// После отработки функции main() отписка и закрытие соединения с сервером NATS-streaming
	defer subscription.Unsubscribe()
	defer sc.Close()

	// Вставляем задержку, чтобы программа оставалась активной
	time.Sleep(120 * time.Second)

}

func startHTTPServer() {
	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		fmt.Println("Ошибка при запуске HTTP сервера:", err)
	}
	fmt.Println("HTTP сервер запущен")
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	// Чтение тела запроса
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка при чтении тела запроса", http.StatusInternalServerError)
		return
	}

	// Вывод содержимого тела запроса в консоль
	fmt.Println("Получен POST-запрос:")
	fmt.Println(string(body))

	// Отправка ответа клиенту
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Запрос успешно обработан"))

	publishMessage(body, channelName, sc)
}

// Функция для отправки сообщения в канал NATS Streaming
func publishMessage(message []byte, channelName string, sc stan.Conn) {

	err := sc.Publish(channelName, message)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Опубликовано сообщение на канале '%s': %s\n", channelName, message)
}
