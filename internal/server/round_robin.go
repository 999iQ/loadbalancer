package server

import (
	"fmt"
	"log"
	"net/http"
)

// BalanceRequestRoundRobin - распределитель запросов по серверам (Round-Robin)
func (lb *LoadBalancer) BalanceRequestRoundRobin(w http.ResponseWriter, r *http.Request) {
	countBackends := lb.pool.GetLenBackends()

	for i := 0; i < countBackends; i++ {
		peer := lb.pool.Next() // Выбираем новый сервер (Round-Robin)
		if !peer.IsAlive() {
			lastErr := fmt.Errorf("failed connection %s -> %s - server is dead. Request has been redirected",
				r.RemoteAddr, peer.URL)
			log.Printf(lastErr.Error())
			continue
		}

		// Пробуем переслать запрос
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}

	log.Printf("FATAL-ERROR: ALL BACKEND-SERVERS ARE DOWN!💀")
	http.Error(w, "Sorry, the service is currently unavailable. Please try again later.", http.StatusServiceUnavailable)
}
