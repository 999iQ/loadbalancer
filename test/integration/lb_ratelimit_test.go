package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"loadbalancer/internal/backend"
	"loadbalancer/internal/config"
	"loadbalancer/internal/ratelimiter/bucket"
	"loadbalancer/internal/ratelimiter/middleware"
	"loadbalancer/internal/server"
)

func TestLoadBalancerWithRateLimiter(t *testing.T) {
	// 1. Настройка тестовых серверов с правильными заголовками
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend", "backend1")
		w.WriteHeader(http.StatusOK)
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend", "backend2")
		w.WriteHeader(http.StatusOK)
	}))
	defer backend2.Close()

	// 2. Создаем тестовую конфигурацию
	cfg := &config.Config{
		Port:                     8080,
		ServerShutdownTimeoutSec: 5 * time.Second,
		LBMethod:                 "RR",
		Backends: []string{
			backend1.URL,
			backend2.URL,
		},
		RateLimit: struct {
			Enabled         bool          `yaml:"enabled"`
			CleanupInterval time.Duration `yaml:"cleanup_interval"`
			Default         struct {
				RequestsPerSec int `yaml:"requests_per_sec"`
				Burst          int `yaml:"burst"`
			} `yaml:"default"`
			SpecialLimits []struct {
				IPs   []string `yaml:"ips"`
				Limit struct {
					RequestsPerSec int `yaml:"requests_per_sec"`
					Burst          int `yaml:"burst"`
				} `yaml:"limit"`
			} `yaml:"special_limits"`
		}{
			Enabled:         true,
			CleanupInterval: 1 * time.Minute,
			Default: struct {
				RequestsPerSec int `yaml:"requests_per_sec"`
				Burst          int `yaml:"burst"`
			}{RequestsPerSec: 10, Burst: 20},
		},
	}

	// 3. Инициализация бакет менеджера
	backendPool := backend.NewPool(cfg.Backends)
	bm := bucket.NewBucketManager(cfg)
	defer bm.Stop()

	// 4. Создаем тестовый load balancer
	lb := server.NewLoadBalancer(cfg.Port, backendPool)

	// 5. Создаем тестовый HTTP сервер
	var handler http.HandlerFunc
	if cfg.LBMethod == "RR" {
		handler = http.HandlerFunc(lb.BalanceRequestRoundRobin)
	} else {
		handler = http.HandlerFunc(lb.BalanceRequestLeastConns)
	}

	testServer := httptest.NewServer(
		middleware.RateLimitMiddleware(bm, handler),
	)
	defer testServer.Close()

	// 6. Тестируем балансировку
	t.Run("Load balancing", func(t *testing.T) {
		backendsHit := make(map[string]int)
		var mu sync.Mutex

		// Делаем 10 запросов с разными IP
		for i := 0; i < 10; i++ {
			req, err := http.NewRequest("GET", testServer.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.RemoteAddr = "192.168.1." + strconv.Itoa(i+1) + ":12345"

			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse // Не следовать редиректам
				},
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			// Проверяем что запрос был перенаправлен на бэкенд
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
				continue
			}

			backendID := resp.Header.Get("X-Backend")
			if backendID == "" {
				t.Error("Missing X-Backend header in response")
				continue
			}

			mu.Lock()
			backendsHit[backendID]++
			mu.Unlock()
		}

		// Проверяем что запросы распределились между бэкендами
		if len(backendsHit) != 2 {
			t.Errorf("Expected requests to be distributed between 2 backends, got %d", len(backendsHit))
		} else {
			t.Logf("Requests distribution: %v", backendsHit)
		}
	})

	time.Sleep(1 * time.Second) // Задержка для восполнения лимитов запросов

	// 7. Тестируем rate limiting
	t.Run("Rate limiting", func(t *testing.T) {
		// Тест с одним IP для проверки лимитов
		req, err := http.NewRequest("GET", testServer.URL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.RemoteAddr = "10.0.0.1:12345" // Фиксированный IP

		// Делаем 30 запросов
		var successCount, rateLimitedCount int
		for i := 0; i < 30; i++ {
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Logf("Request failed: %v", err)
				continue
			}
			resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusOK:
				successCount++
			case http.StatusTooManyRequests:
				rateLimitedCount++
			}
		}

		// Проверяем что первые 20 запросов прошли, остальные отклонены
		if successCount != 20 {
			t.Errorf("Expected 20 successful requests, got %d", successCount)
		}
		if rateLimitedCount != 10 {
			t.Errorf("Expected 10 rate-limited requests, got %d", rateLimitedCount)
		}
	})

	// 8. Тестируем graceful shutdown
	t.Run("Graceful shutdown", func(t *testing.T) {
		// Создаем отдельный сервер для этого теста
		lb := server.NewLoadBalancer(cfg.Port, backend.NewPool(cfg.Backends))
		bm := bucket.NewBucketManager(cfg)
		defer bm.Stop()

		srv := &http.Server{
			Addr:    ":" + strconv.Itoa(cfg.Port),
			Handler: middleware.RateLimitMiddleware(bm, http.HandlerFunc(lb.BalanceRequestRoundRobin)),
		}

		// Запускаем сервер в горутине
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				t.Errorf("Server error: %v", err)
			}
		}()

		// Даем серверу время на запуск
		time.Sleep(100 * time.Millisecond)

		// Имитируем сигнал завершения
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ServerShutdownTimeoutSec*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			t.Errorf("Graceful shutdown failed: %v", err)
		}
	})
}
