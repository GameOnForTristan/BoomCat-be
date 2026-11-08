package middleware

import (
	"net/http"
	"strings"
)

// corsConfig CORS 配置
type corsConfig struct {
	allowedOrigins    []string
	allowedMethods    []string
	allowedHeaders    []string
	allowCredentials bool
	maxAge            string
}

// defaultCORSConfig 默认 CORS 配置（开发环境）
func defaultCORSConfig() corsConfig {
	return corsConfig{
		allowedOrigins:    []string{"*"},
		allowedMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		allowedHeaders:    []string{"Content-Type", "Authorization", "X-Requested-With", "Accept"},
		allowCredentials: false,
		maxAge:            "86400",
	}
}

// CORS 返回 CORS 跨域中间件
func CORS(next http.Handler) http.Handler {
	return CORSWithConfig(defaultCORSConfig())(next)
}

// CORSWithConfig 使用自定义配置创建 CORS 中间件
func CORSWithConfig(cfg corsConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			allowed := false
			for _, o := range cfg.allowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if cfg.allowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.allowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.allowedHeaders, ", "))
				w.Header().Set("Access-Control-Max-Age", cfg.maxAge)
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
