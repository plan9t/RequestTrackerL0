package main

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/stan.go"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Название канала ля NATS-STREAMING
var channelName = "WBChannel"

// Идентификаторы кластера и клиента для NATS-STREAMING
var clusterID = "WB-cluster"
var clientID = "plan9t-client"

// URL NATS Streaming сервера
var natsURL = "nats://localhost:4222"

// Экземпляр подключения к NATS Streaming
var sc stan.Conn

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

	// запуск сервера для интерфейса
	http.HandleFunc("/", IndexHandler)
	go http.ListenAndServe(":4444", nil)

	// Восстановление кэша
	MyCache.AddOrders(GetOrdersFromPostgreSQL())

	// Проверка подключения к серверу NATS-streaming
	if sc.NatsConn().IsConnected() {
		fmt.Println("ПОДКЛЮЧЕНО К СЕРВЕРУ NATS-STREAMING")
	} else {
		fmt.Println("Не подключено к серверу NATSSTREAMING")
	}

	// Подписка на канал и прослушка канала (здесь мы парсим JSON и записываем данные в PostgreSQL
	_, err := createDurableSubscription(sc, channelName, clientID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Текущий кэш:")
	fmt.Println(MyCache.Orders)
	//
	//fmt.Println("ИСКОМЫЙ ЗАКАЗ, КОТОРОГО НЕТ!! В КЭШЕ")
	//fmt.Println(MyCache.GetOrderFromCacheOrDB("5__2363feb7b2b84b6t13371337"))
	//
	//fmt.Println("Обновленный кэш:")
	//fmt.Println(MyCache.Orders)

	// Создание канала для ожидания завершения работы программы
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Ожидание сигнала завершения работы
	select {
	case <-done:
		// Закрытие подключения к NATS Streaming

		sc.Close()

		// вывод кэша
		fmt.Println("Кэш перед выходом из сервиса")
		fmt.Println(MyCache.Orders)
		fmt.Println("Программа завершена.")

	}

}

func createDurableSubscription(sc stan.Conn, channelName, clientID string) (stan.Subscription, error) {
	subscription, err := sc.Subscribe(channelName, func(msg *stan.Msg) {
		fmt.Printf("Получено сообщение из канала '%s': %s\n", channelName, string(msg.Data))

		// Вызрв функции для парсинга JSON и сохранения в базу данных
		orderData, err := parseJSON(msg.Data)
		if err != nil {
			fmt.Println("Ошибка при парсинге JSON:", err)
			return
		}
		fmt.Println("JSON спашрен")

		err = SaveToPostgreSQL(orderData)
		if err != nil {
			fmt.Println("Ошибка при сохранении в PostgreSQL:", err)
			return
		}
		fmt.Println("Данные из темы ", channelName, " сохранены в БД")
		// Запись в кэш
		MyCache.AddOrder(orderData)
		fmt.Println("Данные из темы ", channelName, " записаны в КЭШ")
	}, stan.DurableName(clientID))

	if err != nil && err == stan.ErrBadSubscription {
		// подписка с таким прочным именем уже существует, отписываемся от нее
		if err := subscription.Unsubscribe(); err != nil {
			return nil, err
		}

		// создаем новую подписку с тем же прочным именем
		subscription, err = sc.Subscribe(channelName, func(msg *stan.Msg) {
			fmt.Printf("Получено сообщение из канала '%s': %s\n", channelName, string(msg.Data))

			// Вызов функции для парсинга JSON и сохранения в базу данных
			orderData, err := parseJSON(msg.Data)
			if err != nil {
				fmt.Println("Ошибка при парсинге JSON:", err)
				return
			}
			fmt.Println("JSON спашрен")

			err = SaveToPostgreSQL(orderData)
			if err != nil {
				fmt.Println("Ошибка при сохранении в PostgreSQL:", err)
				return
			}
			fmt.Println("Данные из темы ", channelName, " сохранены в БД")
			// Запись в кэш
			MyCache.AddOrder(orderData)
			fmt.Println("Данные из темы ", channelName, " записаны в КЭШ")
		}, stan.DurableName(clientID))
	}

	return subscription, err
}

func (c *Cache) GetOrderFromCacheOrDB(orderUID string) []Order {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Поиск заказа в кэше
	for _, order := range c.Orders {
		if order.OrderUID == orderUID {
			return []Order{order}
		}
	}

	// Если заказа нет в кэше, получаем его из базы данных
	orders, err := GetOrderByOrderUID(orderUID)
	if err != nil {
		return nil
	}

	// Добавляем полученные заказы в кэш
	c.Orders = append(c.Orders, orders...)

	return orders
}

type AjaxResponse struct {
	InputData  string `json:"inputData"`
	ResultData string `json:"resultData"`
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, "index", nil)
	} else if r.Method == http.MethodPost {
		// Обработка POST-запроса
		var requestBody struct {
			InputData string `json:"inputData"`
		}

		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		inputData := requestBody.InputData

		//действия с данными

		resultData := string(orderToJSON(MyCache.GetOrderFromCacheOrDB(inputData)))

		//resultData := "tessssst"

		// Отправка данных обратно в формате JSON
		response := AjaxResponse{
			InputData:  inputData,
			ResultData: resultData,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func renderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	tmpl, err := template.New("").ParseGlob("templates/*.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
