// Package router 负责路由注册与中间件组装
package router

import (
	"net/http"

	"boom-cat-be/internal/controller"
	"boom-cat-be/internal/middleware"
)

// Router 主路由，封装 http.ServeMux 与中间件链
type Router struct {
	mux    *http.ServeMux
	chain  []func(http.Handler) http.Handler
}

// New 创建路由器
func New() *Router {
	return &Router{
		mux: http.NewServeMux(),
		chain: []func(http.Handler) http.Handler{
			middleware.Recovery,
			middleware.Logging,
			middleware.CORS,
		},
	}
}

// Handler 返回组装好中间件的 http.Handler
func (r *Router) Handler() http.Handler {
	var h http.Handler = r.mux
	for i := len(r.chain) - 1; i >= 0; i-- {
		h = r.chain[i](h)
	}
	return h
}

// Mount 注册路由表
func (r *Router) Mount(routes *Routes) {
	routes.Register(r.mux)
}

// Mux 返回底层 ServeMux（用于自定义路由注册）
func (r *Router) Mux() *http.ServeMux { return r.mux }

// =====================================================================
// 依赖容器：用于组装 controller
// =====================================================================

// Dependencies 路由依赖
type Dependencies struct {
	AuthCtrl *controller.AuthController
	RoomCtrl *controller.RoomController
	WSCtrl   *controller.WSController
	GameCtrl *controller.GameController
}

// Build 根据 deps 构建完整的路由
func Build(deps *Dependencies) http.Handler {
	r := New()
	routes := &Routes{deps: deps}
	r.Mount(routes)
	return r.Handler()
}
