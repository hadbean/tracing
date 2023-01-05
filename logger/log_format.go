package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLayout 日志layout
type LogLayout struct {
	Time       time.Time
	Metadata   map[string]interface{} // 存储自定义原数据
	Path       string                 // 访问路径
	Method     string                 //访问方法
	Query      string                 // 携带query
	Body       string                 // 携带body数据
	IP         string                 // ip地址
	UserAgent  string                 // 代理
	Error      string                 // 错误
	Cost       time.Duration          // 花费时间
	StatusCode int                    //状态码
	Traceid    string                 //事务ID
	Resp       string                 //请求返回内容
}

var log *zap.Logger

type LogConfigs struct {
	LogLevel          string // 日志打印级别
	LogFormat         string // 输出日志格式
	LogPath           string // 输出日志文件路径
	LogFileName       string // 输出日志文件名称
	LogFileMaxSize    int    // 单个日志文件最多存储量 单位(MB)
	LogFileMaxBackups int    // 日志备份文件最多数量
	LogMaxAge         int    // 日志保留时间 单位(Day)
	LogCompress       bool   // 是否压缩日志
	LogStdout         bool   // 是否输出到控制台
	Source            string //平台来源
}

var logCfg = &LogConfigs{
	LogLevel:          "info",
	LogFormat:         "",
	LogPath:           "./log",
	LogFileName:       "eyebox.log",
	LogFileMaxSize:    2,
	LogFileMaxBackups: 12,
	LogMaxAge:         31,
	LogCompress:       false,
	LogStdout:         true,
	Source:            "platform",
}

func Log() *zap.Logger {
	if log != nil {
		return log
	} else {
		InitLogger(nil)
		return log
	}
}

// InitLogger 初始化
func InitLogger(cfg *LogConfigs) (*zap.Logger, error) {
	if log != nil {
		return log, nil
	}
	if cfg != nil {
		logCfg = cfg
	}
	logLevel := map[string]zapcore.Level{
		"debug": zapcore.DebugLevel,
		"info":  zapcore.InfoLevel,
		"warn":  zapcore.WarnLevel,
		"error": zapcore.ErrorLevel,
	}
	writeSyncer, err := getLogWriter()
	if err != nil {
		return nil, err
	}
	encoder := getZapEncoder()
	level, ok := logLevel[logCfg.LogLevel]
	if !ok {
		level = logLevel["info"]
	}
	core := zapcore.NewCore(encoder, writeSyncer, level)
	log = zap.New(core, zap.Hooks(func(e zapcore.Entry) error {
		return nil
	}), zap.AddCaller())
	log = zap.New(core, zap.Hooks(func(e zapcore.Entry) error {
		return nil
	}), zap.AddCaller())
	zap.ReplaceGlobals(log)
	return log, nil
}

// getZapEncoder 获取编码器
func getZapEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.TimeKey = "created_at" //方便json输出是
	encoderConfig.EncodeCaller = func(ec zapcore.EntryCaller, pae zapcore.PrimitiveArrayEncoder) {
		pae.AppendString(fmt.Sprintf("%s  [%s][%s]", ec.TrimmedPath(), logCfg.Source, Goid()))
	}
	if logCfg.LogFormat == "json" {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// getLogWriter 获取日志输出
func getLogWriter() (zapcore.WriteSyncer, error) {
	if strings.Trim(logCfg.LogPath, " \t") == "" {
		return nil, fmt.Errorf("日志路径未配置")
	}
	if strings.Trim(logCfg.LogFileName, " \t") == "" {
		return nil, fmt.Errorf("日志文件名未配置")
	}
	_, err := os.Stat(logCfg.LogPath)
	if err != nil && !os.IsExist(err) {
		if err = os.MkdirAll(logCfg.LogPath, 0o755); err != nil {
			return nil, fmt.Errorf("创建日志目录[%s]失败: %s", logCfg.LogPath, err.Error())
		}
	}
	ljLogger := &lumberjack.Logger{
		Filename:   filepath.Join(logCfg.LogPath, logCfg.LogFileName), // 日志文件路径
		MaxSize:    logCfg.LogFileMaxSize,                             // 单个日志文件大小
		MaxBackups: logCfg.LogFileMaxBackups,                          // 日志备份数量
		MaxAge:     logCfg.LogMaxAge,                                  // 日志最长保留时间
		Compress:   logCfg.LogCompress,                                // 是否压缩日志
	}
	if logCfg.LogStdout {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(ljLogger), zapcore.AddSync(os.Stdout)), nil
	} else {
		return zapcore.AddSync(ljLogger), nil
	}
}
