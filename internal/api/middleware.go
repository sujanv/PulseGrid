package api

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterInterceptor) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w *http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		r.Header.Set("X-Request-ID", reqID)

		interceptor := &responseWriterInterceptor{ResponseWriter: *w, statusCode: http.StatusOK}
		*w = interceptor

		interceptor.Header().Set("X-Request-ID", reqID)

		next.ServeHTTP(*w, r)

		log.Printf("[HTTP] %s %s | %d | %v | req_id=%s",
			r.Method, r.URL.Path, interceptor.statusCode, time.Since(start), reqID)
	})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w *http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[PANIC RECOVERED] %v\n%s", rec, debug.Stack())
				ErrorResponse(*w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(*w, r)
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w *http.ResponseWriter, r *http.Request) {
		(*w).Header().Set("Access-Control-Allow-Origin", "*")
		(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

		if r.Method == http.MethodOptions {
			(*w).WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(*w, r)
	})
}
