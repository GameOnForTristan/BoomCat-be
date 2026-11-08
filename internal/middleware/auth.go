package middleware

// auth.go 占位文件 - JWT 认证中间件
//
// 注意：本任务按用户要求暂不实现 JWT 认证中间件。
// 认证逻辑当前直接在 controller 层通过 AuthService.ValidateToken 完成。
// 如需启用中间件方式，可参考下方接口签名实现：

// import (
// 	"net/http"
// 	"strings"
//
// 	"boom-cat-be/internal/service"
// 	"boom-cat-be/internal/view"
// )
//
// func Auth(authSvc service.AuthService) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			token := extractBearerToken(r)
// 			if token == "" {
// 				view.WriteUnauthorized(w, "missing token")
// 				return
// 			}
// 			_, err := authSvc.ValidateToken(token)
// 			if err != nil {
// 				view.WriteUnauthorized(w, err.Error())
// 				return
// 			}
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }
//
// func extractBearerToken(r *http.Request) string {
// 	auth := r.Header.Get("Authorization")
// 	const prefix = "Bearer "
// 	if !strings.HasPrefix(auth, prefix) {
// 		return ""
// 	}
// 	return strings.TrimPrefix(auth, prefix)
// }
