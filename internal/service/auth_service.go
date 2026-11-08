package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"boom-cat-be/internal/config"
	"boom-cat-be/internal/repository"
)

// AuthResult 认证结果，包含令牌与用户信息
type AuthResult struct {
	UserID       string    `json:"user_id"`
	Nickname     string    `json:"nickname"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Claims JWT 自定义 claims
type Claims struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	jwt.RegisteredClaims
}

// AuthService 认证服务接口
type AuthService interface {
	Register(ctx context.Context, nickname, password string) (*AuthResult, error)
	Login(ctx context.Context, userID, password string) (*AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (*AuthResult, error)
	ValidateToken(token string) (*Claims, error)
	Logout(ctx context.Context, accessToken string) error
}

type authService struct {
	userRepo    repository.UserRepository
	tokenRepo   repository.TokenRepository
	sessionRepo repository.SessionRepository
	jwtSecret   []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
	bcryptCost  int
}

// NewAuthService 创建认证服务
func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	sessionRepo repository.SessionRepository,
	cfg config.AuthConfig,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		sessionRepo: sessionRepo,
		jwtSecret:   []byte(cfg.JWTSecret),
		accessTTL:   time.Duration(cfg.AccessTokenTTL) * time.Second,
		refreshTTL:  time.Duration(cfg.RefreshTokenTTL) * time.Second,
		bcryptCost:  cfg.BcryptCost,
	}
}

// validateNickname 校验昵称：2-16 字符
func validateNickname(nickname string) error {
	n := strings.TrimSpace(nickname)
	if len(n) < 2 || len(n) > 16 {
		return errors.New("nickname must be 2-16 characters")
	}
	return nil
}

// validatePassword 校验密码：6-32 字符
func validatePassword(password string) error {
	if len(password) < 6 || len(password) > 32 {
		return errors.New("password must be 6-32 characters")
	}
	return nil
}

// Register 注册新用户：bcrypt 哈希后写入 PostgreSQL，并签发 JWT
func (s *authService) Register(ctx context.Context, nickname, password string) (*AuthResult, error) {
	if err := validateNickname(nickname); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt hash: %w", err)
	}

	userID := uuid.NewString()
	now := time.Now()
	user := &repository.User{
		ID:        userID,
		Nickname:  strings.TrimSpace(nickname),
		Password:  string(hash),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicateNickname) {
			return nil, errors.New("nickname already taken")
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.issueTokens(ctx, userID, user.Nickname)
}

// Login 用户登录：按 user_id + 密码校验，成功后签发 JWT
func (s *authService) Login(ctx context.Context, userID, password string) (*AuthResult, error) {
	if userID == "" || password == "" {
		return nil, errors.New("user_id and password are required")
	}
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("unauthorized")
	}

	return s.issueTokens(ctx, user.ID, user.Nickname)
}

// Refresh 刷新令牌：校验刷新令牌有效性后签发新令牌（旋转刷新）
func (s *authService) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	if refreshToken == "" {
		return nil, errors.New("refresh token is required")
	}
	stored, err := s.tokenRepo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, errors.New("invalid refresh token")
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	if stored == nil {
		return nil, errors.New("invalid refresh token")
	}
	if time.Now().After(stored.ExpiresAt) {
		_ = s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
		return nil, errors.New("refresh token expired")
	}

	user, err := s.userRepo.GetUserByID(ctx, stored.UserID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	if err := s.tokenRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("revoke refresh token: %w", err)
	}

	return s.issueTokens(ctx, user.ID, user.Nickname)
}

// ValidateToken 校验 access token 并返回 claims
func (s *authService) ValidateToken(token string) (*Claims, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// Logout 登出：清除会话缓存
func (s *authService) Logout(ctx context.Context, accessToken string) error {
	if accessToken == "" {
		return nil
	}
	return s.sessionRepo.DeleteSession(ctx, accessToken)
}

// issueTokens 签发 access + refresh token，并缓存会话
func (s *authService) issueTokens(ctx context.Context, userID, nickname string) (*AuthResult, error) {
	now := time.Now()
	accessExpiry := now.Add(s.accessTTL)
	refreshExpiry := now.Add(s.refreshTTL)

	claims := &Claims{
		UserID:   userID,
		Nickname: nickname,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessStr, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshStr := uuid.NewString()
	stored := &repository.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		Token:     refreshStr,
		ExpiresAt: refreshExpiry,
		CreatedAt: now,
	}
	if err := s.tokenRepo.SaveRefreshToken(ctx, stored); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	sessionData := repository.SessionData{
		UserID:   userID,
		Nickname: nickname,
	}
	_ = s.sessionRepo.SaveSession(ctx, accessStr, sessionData, s.accessTTL)

	return &AuthResult{
		UserID:       userID,
		Nickname:     nickname,
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresAt:    accessExpiry,
	}, nil
}
