package server

import (
	"loadbalancer/internal/backend"
	"net/http"
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
