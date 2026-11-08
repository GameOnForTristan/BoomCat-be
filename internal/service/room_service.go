package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"

	"boom-cat-be/internal/model"
	"boom-cat-be/internal/repository"
)

// PlayerInfo 玩家公开信息
type PlayerInfo struct {
	ID        string `json:"id"`
	Nickname  string `json:"nickname"`
	IsReady   bool   `json:"is_ready"`
	IsOwner   bool   `json:"is_owner"`
	Connected bool   `json:"connected"`
}

// RoomDetail 房间详情
type RoomDetail struct {
	ID          string       `json:"room_id"`
	Name        string       `json:"name"`
	OwnerID     string       `json:"owner_id"`
	InviteCode  string       `json:"invite_code"`
	MaxPlayers  int          `json:"max_players"`
	PlayerCount int          `json:"player_count"`
	Status      string       `json:"status"`
	Players     []PlayerInfo `json:"players"`
}

// RoomService 房间服务接口
type RoomService interface {
	CreateRoom(ctx context.Context, ownerID, ownerNickname, name string, maxPlayers int) (*RoomDetail, error)
	JoinRoom(ctx context.Context, inviteCode, playerID, nickname string) (*RoomDetail, error)
	LeaveRoom(ctx context.Context, roomID, playerID string) error
	ToggleReady(ctx context.Context, roomID, playerID string) (bool, error)
	StartGame(ctx context.Context, roomID, playerID string) error
	GetRoom(ctx context.Context, roomID string) (*RoomDetail, error)
	DispatchMessage(roomID string, msg model.Message) error
	GetRoomEntity(roomID string) (*model.Room, bool)
}

type roomService struct {
	mu           sync.RWMutex
	rooms        map[string]*model.Room
	inviteToRoom map[string]string

	roomMeta    repository.RoomMetaRepository
	inviteRepo  repository.InviteCodeRepository
	playersRepo repository.RoomPlayersRepository
	gameService GameService

	defaultMaxPlayers int
	minPlayers         int
	maxPlayers         int
}

// NewRoomService 创建房间服务（不直接依赖 GameService，通过 SetGameService 注入）
func NewRoomService(
	roomMeta repository.RoomMetaRepository,
	inviteRepo repository.InviteCodeRepository,
	playersRepo repository.RoomPlayersRepository,
	minPlayers, maxPlayers int,
) RoomService {
	return &roomService{
		rooms:             make(map[string]*model.Room),
		inviteToRoom:      make(map[string]string),
		roomMeta:          roomMeta,
		inviteRepo:        inviteRepo,
		playersRepo:       playersRepo,
		defaultMaxPlayers: maxPlayers,
		minPlayers:        minPlayers,
		maxPlayers:        maxPlayers,
	}
}

// SetGameService 注入 GameService（用于打破构造时的循环依赖）
func (s *roomService) SetGameService(gs GameService) {
	s.gameService = gs
}

// generateInviteCode 生成 6 位邀请码（数字+字母）
func generateInviteCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// generateUniqueInviteCode 生成不冲突的邀请码
func (s *roomService) generateUniqueInviteCode(ctx context.Context) string {
	for i := 0; i < 8; i++ {
		code := generateInviteCode()
		if _, exists := s.inviteToRoom[code]; !exists {
			return code
		}
	}
	return generateInviteCode() + fmt.Sprintf("%d", time.Now().UnixNano()%100)
}

// toDetail 将 Room 实体转换为 RoomDetail
func (s *roomService) toDetail(r *model.Room) *RoomDetail {
	players := make([]PlayerInfo, 0, len(r.Players))
	for _, p := range r.Players {
		players = append(players, PlayerInfo{
			ID:        p.ID,
			Nickname:  p.Nickname,
			IsReady:   p.IsReady,
			IsOwner:   r.IsOwner(p.ID),
			Connected: p.ConnectionStatus == model.ConnectionStatusConnected,
		})
	}
	return &RoomDetail{
		ID:          r.ID,
		Name:        r.Name,
		OwnerID:     r.OwnerID,
		InviteCode:  s.lookupInviteCode(r.ID),
		MaxPlayers:  r.Settings.MaxPlayers,
		PlayerCount: r.PlayerCount(),
		Status:      string(r.Status),
		Players:     players,
	}
}

// lookupInviteCode 反查房间对应的邀请码（内存索引）
func (s *roomService) lookupInviteCode(roomID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for code, rid := range s.inviteToRoom {
		if rid == roomID {
			return code
		}
	}
	return ""
}

// CreateRoom 创建房间
func (s *roomService) CreateRoom(ctx context.Context, ownerID, ownerNickname, name string, maxPlayers int) (*RoomDetail, error) {
	if ownerID == "" {
		return nil, errors.New("owner id is required")
	}
	if maxPlayers <= 0 {
		maxPlayers = s.defaultMaxPlayers
	}
	if maxPlayers < s.minPlayers || maxPlayers > s.maxPlayers {
		return nil, fmt.Errorf("max_players must be between %d and %d", s.minPlayers, s.maxPlayers)
	}
	if name == "" {
		name = ownerNickname + "'s room"
	}

	roomID := uuid.NewString()
	inviteCode := s.generateUniqueInviteCode(ctx)
	settings := model.RoomSettings{
		MaxPlayers:       maxPlayers,
		TurnTimeout:      60,
		NopeWindow:       3,
		ReconnectTimeout: 90,
	}
	room := model.NewRoom(roomID, name, ownerID, settings, time.Now().UnixMilli())
	room.Players = append(room.Players, *model.NewPlayer(ownerID, ownerNickname))

	s.mu.Lock()
	s.rooms[roomID] = room
	s.inviteToRoom[inviteCode] = roomID
	s.mu.Unlock()

	go s.persistRoomMeta(ctx, room, inviteCode)
	go s.playersRepo.AddPlayer(context.Background(), roomID, ownerID)

	return s.toDetail(room), nil
}

// JoinRoom 通过邀请码加入房间
func (s *roomService) JoinRoom(ctx context.Context, inviteCode, playerID, nickname string) (*RoomDetail, error) {
	if inviteCode == "" || playerID == "" {
		return nil, errors.New("invite_code and player_id are required")
	}
	s.mu.RLock()
	roomID, ok := s.inviteToRoom[inviteCode]
	s.mu.RUnlock()
	if !ok {
		return nil, errors.New("invalid invite code")
	}

	s.mu.Lock()
	room, exists := s.rooms[roomID]
	s.mu.Unlock()
	if !exists {
		return nil, errors.New("room not found")
	}

	if room.Status != model.RoomStatusWaiting {
		return nil, model.ErrRoomAlreadyStarted
	}
	if room.PlayerCount() >= room.Settings.MaxPlayers {
		return nil, model.ErrRoomFull
	}
	if _, ok := room.FindPlayer(playerID); ok {
		return s.toDetail(room), nil
	}

	room.Players = append(room.Players, *model.NewPlayer(playerID, nickname))

	go s.persistRoomMeta(ctx, room, inviteCode)
	go s.playersRepo.AddPlayer(context.Background(), roomID, playerID)

	return s.toDetail(room), nil
}

// LeaveRoom 玩家退出房间
func (s *roomService) LeaveRoom(ctx context.Context, roomID, playerID string) error {
	s.mu.Lock()
	room, exists := s.rooms[roomID]
	s.mu.Unlock()
	if !exists {
		return errors.New("room not found")
	}

	if room.Status == model.RoomStatusPlaying {
		return model.ErrRoomAlreadyStarted
	}

	s.mu.Lock()
	idx := -1
	for i, p := range room.Players {
		if p.ID == playerID {
			idx = i
			break
		}
	}
	if idx < 0 {
		s.mu.Unlock()
		return model.ErrPlayerNotFound
	}
	room.Players = append(room.Players[:idx], room.Players[idx+1:]...)

	// 房主退出后转交给第一位玩家
	if room.OwnerID == playerID && len(room.Players) > 0 {
		room.OwnerID = room.Players[0].ID
	}

	empty := len(room.Players) == 0
	inviteCode := s.lookupInviteCode(roomID)
	if empty {
		delete(s.rooms, roomID)
		if inviteCode != "" {
			delete(s.inviteToRoom, inviteCode)
		}
	}
	s.mu.Unlock()

	go func() {
		bg := context.Background()
		s.playersRepo.RemovePlayer(bg, roomID, playerID)
		if empty {
			s.roomMeta.DeleteRoomMeta(bg, roomID)
			if inviteCode != "" {
				s.inviteRepo.DeleteInviteCode(bg, inviteCode)
			}
		} else {
			s.persistRoomMeta(bg, room, inviteCode)
		}
	}()

	return nil
}

// ToggleReady 切换玩家准备状态，返回新的准备状态
func (s *roomService) ToggleReady(ctx context.Context, roomID, playerID string) (bool, error) {
	s.mu.Lock()
	room, exists := s.rooms[roomID]
	s.mu.Unlock()
	if !exists {
		return false, errors.New("room not found")
	}
	if room.Status != model.RoomStatusWaiting {
		return false, model.ErrRoomAlreadyStarted
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := room.FindPlayer(playerID)
	if !ok {
		return false, model.ErrPlayerNotFound
	}
	p.IsReady = !p.IsReady
	return p.IsReady, nil
}

// StartGame 房主开始游戏
func (s *roomService) StartGame(ctx context.Context, roomID, playerID string) error {
	s.mu.Lock()
	room, exists := s.rooms[roomID]
	s.mu.Unlock()
	if !exists {
		return errors.New("room not found")
	}
	if !room.IsOwner(playerID) {
		return model.ErrNotOwner
	}
	if room.Status != model.RoomStatusWaiting {
		return model.ErrRoomAlreadyStarted
	}
	if !room.AllReady() {
		return model.ErrNotAllReady
	}

	if err := s.gameService.InitializeGame(ctx, room); err != nil {
		return err
	}

	room.Status = model.RoomStatusPlaying
	go s.persistRoomMeta(ctx, room, s.lookupInviteCode(roomID))

	// 启动房间消息循环
	go s.gameService.StartRoomLoop(roomID)
	return nil
}

// GetRoom 获取房间详情
func (s *roomService) GetRoom(ctx context.Context, roomID string) (*RoomDetail, error) {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()
	if !exists {
		return nil, errors.New("room not found")
	}
	return s.toDetail(room), nil
}

// DispatchMessage 将玩家消息推送到房间 Inbox
func (s *roomService) DispatchMessage(roomID string, msg model.Message) error {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()
	if !exists {
		return errors.New("room not found")
	}
	select {
	case room.Inbox <- msg:
		return nil
	default:
		return errors.New("room inbox full")
	}
}

// GetRoomEntity 获取房间实体（供 GameService 使用）
func (s *roomService) GetRoomEntity(roomID string) (*model.Room, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	room, ok := s.rooms[roomID]
	return room, ok
}

// persistRoomMeta 异步持久化房间元数据到 Redis
func (s *roomService) persistRoomMeta(ctx context.Context, room *model.Room, inviteCode string) {
	if s.roomMeta == nil {
		return
	}
	bg := context.Background()
	meta := repository.RoomMeta{
		Name:        room.Name,
		OwnerID:     room.OwnerID,
		Status:      string(room.Status),
		MaxPlayers:  room.Settings.MaxPlayers,
		PlayerCount: room.PlayerCount(),
		CreatedAt:   room.CreatedAt,
	}
	_ = s.roomMeta.SaveRoomMeta(bg, room.ID, meta)
	if inviteCode != "" {
		_ = s.inviteRepo.SaveInviteCode(bg, inviteCode, room.ID)
	}
}
