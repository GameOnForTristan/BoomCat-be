package router

import (
	"net/http"

	"boom-cat-be/internal/controller"
)

// Routes 路由表定义
type Routes struct {
	deps *Dependencies
}

// Register 注册所有路由到 mux
func (r *Routes) Register(mux *http.ServeMux) {
	// =====================================================================
	// HTTP RESTful API
	// =====================================================================

	// 认证：/api/v1/auth/*
	mux.HandleFunc("/api/v1/auth/register", r.authRegister)
	mux.HandleFunc("/api/v1/auth/login", r.authLogin)
	mux.HandleFunc("/api/v1/auth/refresh", r.authRefresh)
	mux.HandleFunc("/api/v1/auth/logout", r.authLogout)

	// 房间：/api/v1/rooms （POST 创建）+ /api/v1/rooms/{room_id}/{action} （前缀匹配）
	mux.HandleFunc("/api/v1/rooms", r.roomsRoot)
	mux.HandleFunc("/api/v1/rooms/join", r.roomsJoin)
	mux.HandleFunc("/api/v1/rooms/", r.roomsDynamic)

	// WebSocket
	mux.HandleFunc("/ws", r.websocket)

	// 健康检查
	mux.HandleFunc("/health", r.health)
}

// =====================================================================
// Auth handlers
// =====================================================================

func (r *Routes) authRegister(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.deps.AuthCtrl.Register(w, req)
}

func (r *Routes) authLogin(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.deps.AuthCtrl.Login(w, req)
}

func (r *Routes) authRefresh(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.deps.AuthCtrl.Refresh(w, req)
}

func (r *Routes) authLogout(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.deps.AuthCtrl.Logout(w, req)
}

// =====================================================================
// Room handlers
// =====================================================================

// roomsRoot 处理 /api/v1/rooms （仅 POST 创建）
func (r *Routes) roomsRoot(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.deps.RoomCtrl.Create(w, req)
}

// roomsJoin 处理 /api/v1/rooms/join
func (r *Routes) roomsJoin(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.deps.RoomCtrl.Join(w, req)
}

// roomsDynamic 处理 /api/v1/rooms/{room_id}/{action}
// net/http ServeMux 不支持路径参数，手动解析
func (r *Routes) roomsDynamic(w http.ResponseWriter, req *http.Request) {
	action, roomID := parseRoomPath(req.URL.Path)
	if roomID == "" {
		http.NotFound(w, req)
		return
	}

	// 将 room_id 注入 query 供 controller 使用
	query := req.URL.Query()
	if query.Get("room_id") == "" {
		query.Set("room_id", roomID)
		req.URL.RawQuery = query.Encode()
	}

	switch action {
	case "leave":
		if req.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.deps.RoomCtrl.Leave(w, req)
	case "ready":
		if req.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.deps.RoomCtrl.ToggleReady(w, req)
	case "start":
		if req.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.deps.RoomCtrl.Start(w, req)
	case "state":
		if req.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.deps.GameCtrl.GetState(w, req)
	case "event":
		if req.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.deps.GameCtrl.SendEvent(w, req)
	case "":
		// GET /api/v1/rooms/{room_id}
		if req.Method == http.MethodGet {
			r.deps.RoomCtrl.Get(w, req)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	default:
		http.NotFound(w, req)
	}
}

// =====================================================================
// WebSocket handler
// =====================================================================

func (r *Routes) websocket(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.deps.WSCtrl.HandleWS(w, req)
}

// =====================================================================
// Health check
// =====================================================================

func (r *Routes) health(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// =====================================================================
// 辅助函数
// =====================================================================

// parseRoomPath 从 path /api/v1/rooms/{room_id}/{action} 中解析 room_id 与 action
// 返回 (action, roomID)，无法解析时 roomID 为空
// 例如：/api/v1/rooms/abc123/leave -> (action="leave", roomID="abc123")
//      /api/v1/rooms/abc123       -> (action="", roomID="abc123")
func parseRoomPath(path string) (action, roomID string) {
	const prefix = "/api/v1/rooms/"
	if len(path) <= len(prefix) {
		return "", ""
	}
	rest := path[len(prefix):]
	for i, c := range rest {
		if c == '/' {
			return rest[i+1:], rest[:i]
		}
	}
	return "", rest
}

// 编译期确保 Dependencies 字段类型存在
var _ *controller.AuthController = (*controller.AuthController)(nil)
