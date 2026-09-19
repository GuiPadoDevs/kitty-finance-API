package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriterInterceptor intercepta o status code da resposta HTTP
type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// RequestLogger faz o log estruturado e elegante das requisições recebidas
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		interceptor := &responseWriterInterceptor{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(interceptor, r)

		duration := time.Since(start)
		log.Printf("🎀 [%s] %s -> %d (%s)", r.Method, r.URL.Path, interceptor.statusCode, duration)
	})
}
