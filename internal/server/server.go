package server

import (
	"context"
	"fmt"
	"loadbalancer/internal/backend"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type LoadBalancer struct {
	port   int // порт балансировщика (по дефолту 8080)
	pool   *backend.Pool
	server *http.Server // для shutdown
}

// NewLoadBalancer - конструктор для объекта LoadBalancer
func NewLoadBalancer(port int, pool *backend.Pool) *LoadBalancer {
	return &LoadBalancer{
		port: port,
		pool: pool,
	}
}

func (lb *LoadBalancer) Start() error {
	lb.server = &http.Server{
		Addr:    ":" + strconv.Itoa(lb.port),
		Handler: http.HandlerFunc(lb.balanceRequest),
	}

	// Канал для обработки сигналов завершения
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // убрать в конфиг
	defer cancel()

	if err := lb.server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
		return err
	}

	log.Println("Server gracefully stopped")
	return nil
}

func (lb *LoadBalancer) balanceRequest(w http.ResponseWriter, r *http.Request) {
	maxRetries := 3                      // убрать в конфиг
	retryDelay := 100 * time.Millisecond // Задержка между попытками // убрать в конфиг

	var lastErr error

	for i := 0; i < maxRetries; i++ {
		peer := lb.pool.Next()
		if !peer.IsAlive() {
			lastErr = fmt.Errorf("backend %s is not alive", peer.URL.String())
			time.Sleep(retryDelay)
			continue
		}

		// Пробуем переслать запрос
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}

	// Все попытки исчерпаны
	http.Error(w, "Service unavailable: "+lastErr.Error(), http.StatusServiceUnavailable)
}
