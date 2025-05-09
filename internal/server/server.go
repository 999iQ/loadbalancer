package server

import (
	"context"
	"loadbalancer/internal/config"
	"loadbalancer/internal/errors/errors_middleware"
	"loadbalancer/internal/ratelimiter/bucket"
	"loadbalancer/internal/ratelimiter/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// StartServer - запускает сервер с балансировщиком
func (lb *LoadBalancer) StartServer(conf *config.Config) error {

	// Инициализирую бакет менеджер для Rate Limiter
	bm := bucket.NewBucketManager(conf)
	defer bm.Stop()

	// BalanceMethod - спец. тип чтобы можно было передать метод балансировки из конфига
	type BalanceMethod func(w http.ResponseWriter, r *http.Request)
	var balanceMethod BalanceMethod
	if conf.LBMethod == "RR" {
		balanceMethod = lb.BalanceRequestRoundRobin
	} else if conf.LBMethod == "LC" {
		balanceMethod = lb.BalanceRequestLeastConns
	}

	// заворачиваем балансировщик в ограничитель и сверху ещё обработчик ошибок
	handler := errors_middleware.ErrorHandler(
		middleware.RateLimitMiddleware(bm, http.HandlerFunc(balanceMethod)))

	// инит сервера с выбором метода loadBalancer'а
	lb.server = &http.Server{
		Addr: ":" + strconv.Itoa(lb.port),
		// ниже описываю errors_middleware и следующий хэндлер для вызова после проверки IP rate limit'ером
		Handler: handler,
	}

	// канал для обработки сигналов завершения программы
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

	// Даём серверу время на завершение активных соединений
	ctx, cancel := context.WithTimeout(context.Background(), conf.ServerShutdownTimeoutSec*time.Second)
	defer cancel()

	// server stop
	if err := lb.server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
		return err
	}

	log.Println("Server gracefully stopped.")
	return nil
}
