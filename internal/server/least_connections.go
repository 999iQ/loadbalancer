package server

import (
	"log"
	"net/http"
)

// BalanceRequestLeastConns - распределитель запросов по серверам (Least Connections)
func (lb *LoadBalancer) BalanceRequestLeastConns(w http.ResponseWriter, r *http.Request) {
	peer := lb.pool.GetLeastBusyBackend()
	if peer == nil { // все мертвы
		log.Printf("FATAL-ERROR: ALL BACKEND-SERVERS ARE DOWN!💀")
		http.Error(w, "Sorry, the service is currently unavailable. Please try again later.", http.StatusServiceUnavailable)
		return
	}
	peer.IncrementConn()
	defer peer.DecrementConn()

	peer.ReverseProxy.ServeHTTP(w, r)
	return
}
