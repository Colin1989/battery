package logger

import (
	"context"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log/slog"
	"os"
	"runtime"
)

var (
	defaultLogger *slog.Logger
	nodeId        string // current node id
)

func init() {
	defaultLogger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func NewLogger(config LogConfig) {
	log := &lumberjack.Logger{
		Filename:   config.LogPath,    // 日志文件的位置
		MaxSize:    config.MaxSize,    // 文件最大尺寸（以MB为单位）
		MaxBackups: config.MaxBackups, // 保留的最大旧文件数量
		MaxAge:     config.MaxAge,     // 保留旧文件的最大天数
		Compress:   true,              // 是否压缩/归档旧文件
		LocalTime:  true,              // 使用本地时间创建时间戳
	}
	defaultLogger = slog.New(slog.NewJSONHandler(io.MultiWriter(log, os.Stderr), &slog.HandlerOptions{}))
	if len(nodeId) != 0 {
		defaultLogger = defaultLogger.WithGroup(nodeId)
	}
}

func Caller(depth int) (fileAndLine slog.Value) {
	if _, file, line, ok := runtime.Caller(depth + 1); ok {
		fileAndLine = slog.StringValue(fmt.Sprintf("%s:%d", file, line))
	}
	return
}

func callerAttr() slog.Attr {
	return slog.String("Caller", Caller(2).String())
}

func Debug(msg string, attr ...slog.Attr) {
	bg := context.Background()
	level := slog.LevelDebug
	if !defaultLogger.Enabled(bg, level) {
		return
	}
	attr = append(attr, callerAttr())
	defaultLogger.LogAttrs(context.Background(), level, msg, attr...)
}

func Infof(msg string, args ...interface{}) {
	bg := context.Background()
	level := slog.LevelInfo
	if !defaultLogger.Enabled(bg, level) {
		return
	}
	defaultLogger.Log(context.Background(), level, fmt.Sprintf(msg, args...))
}

func Info(msg string, attr ...slog.Attr) {
	bg := context.Background()
	level := slog.LevelInfo
	if !defaultLogger.Enabled(bg, level) {
		return
	}
	attr = append(attr, callerAttr())
	defaultLogger.LogAttrs(context.Background(), level, msg, attr...)
}

func Warn(msg string, attr ...slog.Attr) {
	bg := context.Background()
	level := slog.LevelWarn
	if !defaultLogger.Enabled(bg, level) {
		return
	}
	attr = append(attr, callerAttr())
	defaultLogger.LogAttrs(context.Background(), level, msg, attr...)
}

func Error(msg string, attr ...slog.Attr) {
	bg := context.Background()
	level := slog.LevelError
	if !defaultLogger.Enabled(bg, level) {
		return
	}
	attr = append(attr, callerAttr())
	defaultLogger.LogAttrs(context.Background(), level, msg, attr...)
}

func Fatal(msg string, attr ...slog.Attr) {
	bg := context.Background()
	level := slog.LevelError
	if !defaultLogger.Enabled(bg, level) {
		return
	}
	attr = append(attr, callerAttr())
	defaultLogger.LogAttrs(context.Background(), level, msg, attr...)
	os.Exit(1)
}

func CallerStack(err error, skip int) {
	defaultLogger.LogAttrs(context.Background(),
		slog.LevelError,
		err.Error(),
		slog.String("Stack", string(stack(skip+1))))
}
