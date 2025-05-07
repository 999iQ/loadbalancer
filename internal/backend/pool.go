package backend

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	URL          *url.URL
	Alive        bool // флаг доступности сервера
	ReverseProxy *httputil.ReverseProxy
	mux          sync.RWMutex
}

// SetAlive - изменяет статус бэкенда
func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Alive = alive
}

// IsAlive - чтение статуса бэкенда
func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.Alive
}

type Pool struct {
	backends []*Backend
	current  uint64
	mux      sync.RWMutex
}

// NewPool - создаёт пул бэкендов из переданного списка адресов серверов
func NewPool(backendURLs []string) *Pool {
	var pool Pool
	for _, u := range backendURLs {
		parsedURL, _ := url.Parse(u)
		proxy := httputil.NewSingleHostReverseProxy(parsedURL)
		pool.backends = append(pool.backends, &Backend{
			URL:          parsedURL,
			Alive:        true,
			ReverseProxy: proxy,
		})
	}
	return &pool
}

// Next - реализация балансировки методом Round-Robin
func (p *Pool) Next() *Backend {
	p.mux.Lock()
	defer p.mux.Unlock()

	next := int(p.current+1) % len(p.backends)
	p.current = uint64(next)
	return p.backends[next]
}

// HealthCheck - периодично проверяет статусы серверов
func (p *Pool) HealthCheck(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, b := range p.backends {
				resp, err := http.Get(b.URL.String() + "/health")
				alive := err == nil && resp.StatusCode == http.StatusOK
				b.SetAlive(alive)
			}
		case <-ctx.Done(): // Остановка по сигналу или красиво "Graceful Shutdown"
			log.Println("HealthCheck stopped")
			return
		}
	}
}
