// Package utils 提供通用工具函数：ID 生成、邀请码生成、时间工具
package utils

import (
	"crypto/rand"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
)

// =====================================================================
// ID 生成器
// =====================================================================

// NewID 生成一个 UUID v4 字符串
func NewID() string {
	return uuid.NewString()
}

// NewIDWithPrefix 生成带前缀的 ID，例如 room_xxx、player_xxx
func NewIDWithPrefix(prefix string) string {
	return prefix + "_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
}

// NewTraceID 生成追踪 ID（用于请求链路）
func NewTraceID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

// =====================================================================
// 邀请码生成
// =====================================================================

// inviteCharset 邀请码字符集（去除歧义字符 0/O/1/I/L）
const inviteCharset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

// GenerateInviteCode 生成 6 位邀请码
func GenerateInviteCode() string {
	return generateCode(inviteCharset, 6)
}

// GenerateShortCode 生成指定长度的短码
func GenerateShortCode(length int) string {
	if length <= 0 {
		length = 6
	}
	return generateCode(inviteCharset, length)
}

// generateCode 用 crypto/rand 生成安全随机码
func generateCode(charset string, length int) string {
	b := make([]byte, length)
	max := big.NewInt(int64(len(charset)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			// 退化到时间戳避免阻塞
			b[i] = charset[int(time.Now().UnixNano())%len(charset)]
			continue
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// =====================================================================
// 时间工具
// =====================================================================

// NowMillis 当前毫秒级 Unix 时间戳
func NowMillis() int64 { return time.Now().UnixMilli() }

// NowSeconds 当前秒级 Unix 时间戳
func NowSeconds() int64 { return time.Now().Unix() }

// NowTime 当前 time.Time
func NowTime() time.Time { return time.Now() }

// FormatTime 格式化为 ISO 8601 字符串
func FormatTime(t time.Time) string { return t.Format(time.RFC3339) }

// ParseTime 解析 ISO 8601 字符串
func ParseTime(s string) (time.Time, error) { return time.Parse(time.RFC3339, s) }

// DurationSince 返回自指定时间戳以来的持续时间（秒）
func DurationSince(startMillis int64) int {
	return int((time.Now().UnixMilli() - startMillis) / 1000)
}

// =====================================================================
// 字符串工具
// =====================================================================

// DefaultIfEmpty 若 s 为空则返回 defaultVal
func DefaultIfEmpty(s, defaultVal string) string {
	if strings.TrimSpace(s) == "" {
		return defaultVal
	}
	return s
}

// Truncate 截断字符串到指定长度
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// MaskString 用 * 替换部分字符（用于敏感信息脱敏）
func MaskString(s string, keepStart, keepEnd int) string {
	if len(s) <= keepStart+keepEnd {
		return strings.Repeat("*", len(s))
	}
	return s[:keepStart] + strings.Repeat("*", len(s)-keepStart-keepEnd) + s[len(s)-keepEnd:]
}
