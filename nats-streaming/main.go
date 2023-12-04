package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/stan.go"
)

var natsURL = "nats://localhost:4222"
var clusterID = "WB-cluster"

func main() {
	// Запуск HTTP сервера для прослушки запросов
	go startHTTPServer()

	sc, err := stan.Connect(clusterID, "plan9t-streamer", stan.NatsURL(natsURL))
	if err != nil {
		log.Fatal(err)
	}

	if sc.NatsConn().IsConnected() {
		fmt.Println("ПОДКЛЮЧЕНО К СЕРВЕРУ NATS-STREAMING")
	} else {
		fmt.Println("Не подключено к серверу NATSSTREAMING")
	}

	// Перехватчик POST запросов
	http.HandleFunc("/forpost", func(w http.ResponseWriter, r *http.Request) {
		message, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Ошибка при чтении тела запроса", http.StatusInternalServerError)
			return
		}

		// Публикация сообщения из перехватчика POST-запросов в тему на NATS-streaming
		err = sc.Publish("WBChannel", message)
		println("ОПУБЛИКОВАНО")
		if err != nil {
			http.Error(w, "Ошибка при отправке сообщения в NATS Streaming", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Сообщение успешно опубликовано в NATS Streaming"))
	})

	// Создание канала для ожидания завершения работы программы
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Ожидание сигнала завершения работы
	select {
	case <-done:
		// Закрытие подключения к NATS Streaming
		sc.Close()
		fmt.Println("Программа завершена.")
	}

}

func startHTTPServer() {
	fmt.Println("HTTP сервер запущен")
	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		fmt.Println("Ошибка при запуске HTTP сервера:", err)
	}

}
