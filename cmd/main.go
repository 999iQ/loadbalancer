package main

import (
	"context"
	"fmt"
	"loadbalancer/internal/backend"
	"loadbalancer/internal/config"
	"loadbalancer/internal/server"
	"log"
)

func main() {
	conf, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Println(conf, err)

	backendPool := backend.NewPool(conf.Backends)

	// Контекст для корректной остановки программы
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем HealthCheck в горутине
	go backendPool.HealthCheck(ctx)

	lb := server.NewLoadBalancer(conf.Port, backendPool)
	if err := lb.Start(); err != nil {
		log.Fatal(err)
	}

}
