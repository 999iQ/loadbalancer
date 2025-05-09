package backend

import (
	"context"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// Pool - список серверов, а также номер того, куда будет направлен следующий входящий запрос
type Pool struct {
	backends []*Backend
	current  uint32
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
	p.current = uint32(next)
	return p.backends[next]
}

// HealthCheck - периодично проверяет статусы серверов
func (p *Pool) HealthCheck(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
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

// GetMaxRetries - геттер-функция возращает количество серверов из пула
func (p *Pool) GetLenBackends() int {
	p.mux.RLock()
	defer p.mux.RUnlock()
	return len(p.backends)
}

// GetLeastBusyBackend - возращает менее занятый бэкенд
func (p *Pool) GetLeastBusyBackend() *Backend {
	var leastBusy *Backend
	minConns := math.MaxInt32

	p.mux.RLock()
	defer p.mux.RUnlock()

	for _, b := range p.backends {
		// проверяем живой ли бэкенд
		if !b.IsAlive() {
			continue
		}

		connectsCount := b.GetActiveConnects()
		if connectsCount < minConns {
			minConns = connectsCount
			leastBusy = b
		}
	}

	return leastBusy // Может быть nil, если все бэкенды мертвы
}
