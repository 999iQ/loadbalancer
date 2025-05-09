package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// Start - запускает сервер с балансировщиком
func (lb *LoadBalancer) Start(ShutdownTimeoutSec time.Duration) error {

	// инит сервера с выбором метода loadBalancer'а
	lb.server = &http.Server{
		Addr:    ":" + strconv.Itoa(lb.port),
		Handler: http.HandlerFunc(lb.balanceRequestLeastConns),
	}

	// Канал для обработки сигналов завершения программы
	stopChan := make(chan os.Signal, 1)
	// настраиваем прослушивание сигналов завершения в этот канал
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM) // SIGINT|SIGTERM

	// Запуск сервера в горутине
	go func() {
		log.Printf("LoadBalancer started on :%d\n", lb.port)
		if err := lb.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	<-stopChan
	log.Println("Shutting down server...")

	// Даём серверу 5 секунд на завершение активных соединений
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeoutSec*time.Second)
	defer cancel()

	// server stop
	if err := lb.server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
		return err
	}

	log.Println("Server gracefully stopped.")
	return nil
}
