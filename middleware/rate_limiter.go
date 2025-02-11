package middleware

import (
	"ass3_part2/models"
	"encoding/json"
	"golang.org/x/time/rate"
	"net/http"
)

var Limiter = rate.NewLimiter(1, 1) // запрос,сек

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !Limiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			response := models.Response{Status: "fail", Message: "Too Many Requests"}
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(response)
			return
		}
		next.ServeHTTP(w, r)
	})
}
