package backend

import (
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

// Backend - один из списка серверов для получения запросов
type Backend struct {
	URL            *url.URL
	Alive          bool // флаг доступности сервера
	ReverseProxy   *httputil.ReverseProxy
	activeConnects int32 // счетчик активных подключений (для lb-метода leastConnections)
	mux            sync.RWMutex
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

// Далее методы для реализации lb-метода leastConnections

// IncrementConn - увеличивает счетчик активных подключений к бэкенду
func (b *Backend) IncrementConn() {
	atomic.AddInt32(&b.activeConnects, 1)
}

// DecrementConn - уменьшает счетчик активных подключений к бэкенду
func (b *Backend) DecrementConn() {
	atomic.AddInt32(&b.activeConnects, -1)
}

// GetActiveConnects - возращает количество подключений к бэкенду
func (b *Backend) GetActiveConnects() int {
	return int(atomic.LoadInt32(&b.activeConnects))
}
