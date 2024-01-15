package blog

import (
	"log/slog"
	"strings"
)

type LogConfig struct {
	LogLevel   string `json:"log_level"`   // 输出日志等级(debug, info, warn, error)
	LogPath    string `json:"log_path"`    // 日志保存路径
	MaxSize    int    `json:"max_size"`    // 文件切割大小(MB)
	MaxAge     int    `json:"max_age"`     // 最大保留天数(达到限制，则会被清理)
	MaxBackups int    `json:"max_backups"` // 最大备份数量
}

func (l LogConfig) Level() slog.Level {
	level := strings.ToLower(l.LogLevel)
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
