package api

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
}

func (cfg *ApiConfig) MidlewareMetricInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cfg.FileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func HealthCheck(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("OK"))

}

func (cfg *ApiConfig) RequestCount(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	count := fmt.Sprintf("Hits: %v\n", cfg.FileserverHits.Load())
	rw.Write([]byte(count))
}

func (cfg *ApiConfig) Reset(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	cfg.FileserverHits = atomic.Int32{}
}
