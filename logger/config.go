package logger

type LogConfig struct {
	Level      string `json:"level"`       // 输出日志等级
	LogPath    string `json:"log_path"`    // 日志保存路径
	MaxSize    int    `json:"max_size"`    // 文件切割大小(MB)
	MaxAge     int    `json:"max_age"`     // 最大保留天数(达到限制，则会被清理)
	MaxBackups int    `json:"max_backups"` // 最大备份数量
}
