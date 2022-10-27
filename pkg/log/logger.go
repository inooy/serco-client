package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

type name interface {
}

var sugaredLogger *zap.SugaredLogger

func init() {
	// 设置一些基本日志格式 具体含义还比较好理解，直接看zap源码也不难懂
	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,
		TimeKey:     "ts",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		CallerKey:    "file",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	})

	// 实现两个判断日志等级的interface
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.InfoLevel
	})

	// 最后创建具体的Logger
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), infoLevel), //打印到控制台
	)

	log := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)) // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 有点小坑
	sugaredLogger = log.Sugar()
}

func SetupLogger(log *zap.SugaredLogger) {
	sugaredLogger = log
}

func Debug(args ...interface{}) {
	sugaredLogger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	sugaredLogger.Debugf(template, args...)
}

func Info(args ...interface{}) {
	sugaredLogger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	sugaredLogger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	sugaredLogger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	sugaredLogger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	sugaredLogger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	sugaredLogger.Errorf(template, args...)
}

func DPanic(args ...interface{}) {
	sugaredLogger.DPanic(args...)
}

func DPanicf(template string, args ...interface{}) {
	sugaredLogger.DPanicf(template, args...)
}

func Panic(args ...interface{}) {
	sugaredLogger.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	sugaredLogger.Panicf(template, args...)
}

func Fatal(args ...interface{}) {
	sugaredLogger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	sugaredLogger.Fatalf(template, args...)
}
