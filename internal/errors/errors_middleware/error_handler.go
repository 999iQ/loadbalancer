package errors_middleware

import (
	"loadbalancer/internal/errors"
	"net/http"
)

func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		defer func() {
			if r := recover(); r != nil {
				handleError(w, errors.NewAPIError(http.StatusInternalServerError, "Internal server error"))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func handleError(w http.ResponseWriter, apiError *errors.APIError) {
	w.WriteHeader(apiError.Code)
	w.Write(apiError.ToJSON())
}
