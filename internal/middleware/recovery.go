package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"boom-cat-be/pkg/logger"
)

// Recovery Panic 恢复中间件，捕获 panic 并返回 500
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := debug.Stack()
				logger.Default.Error("panic recovered",
					logger.F("error", fmt.Sprintf("%v", rec)),
					logger.F("stack", string(stack)),
					logger.F("path", r.URL.Path),
					logger.F("method", r.Method),
				)

				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"code":500,"message":"internal server error","data":null}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
