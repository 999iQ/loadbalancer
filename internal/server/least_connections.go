package server

import (
	"loadbalancer/internal/errors"
	"log"
	"net/http"
)

// BalanceRequestLeastConns - распределитель запросов по серверам (Least Connections)
func (lb *LoadBalancer) BalanceRequestLeastConns(w http.ResponseWriter, r *http.Request) {
	peer := lb.pool.GetLeastBusyBackend()
	if peer == nil { // все мертвы
		log.Printf("FATAL-ERROR: ALL BACKEND-SERVERS ARE DOWN!💀")
		err := errors.NewAPIError(http.StatusServiceUnavailable, "Sorry, the service is currently unavailable. Please try again later.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(err.Code)
		w.Write(err.ToJSON())
		return
	}
	peer.IncrementConn()
	defer peer.DecrementConn()

	peer.ReverseProxy.ServeHTTP(w, r)
	return
}
