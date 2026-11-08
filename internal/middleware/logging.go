package middleware

import (
	"net/http"
	"time"

	"boom-cat-be/pkg/logger"
)

// responseRecorder 记录响应状态码与字节数
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// Logging 请求日志中间件
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rr := &responseRecorder{ResponseWriter: w}

		next.ServeHTTP(rr, r)

		duration := time.Since(start)
		logger.Default.Info("http request",
			logger.F("method", r.Method),
			logger.F("path", r.URL.Path),
			logger.F("status", rr.statusCode),
			logger.F("bytes", rr.bytes),
			logger.F("duration_ms", duration.Milliseconds()),
			logger.F("remote", r.RemoteAddr),
		)
	})
}
