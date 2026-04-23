package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLogger creates an instance of [log.Logger] based on [zap.Logger] implementation.
func NewZapLogger(cfg zap.Config) (Logger, error) {
	l, err := cfg.Build(zap.AddCallerSkip(2)) // 1 for loggerImpl.Log, 1 for EventBuilder.Msg/Send
	if err != nil {
		return nil, err
	}
	return &baseLogger{impl: &zapLogger{logger: l}}, nil
}

func convertLevelToZap(lvl Level) zapcore.Level {
	var zl zapcore.Level
	switch lvl {
	case LevelFatal:
		zl = zapcore.FatalLevel
	case LevelError:
		zl = zapcore.ErrorLevel
	case LevelInfo:
		zl = zapcore.InfoLevel
	case LevelDebug:
		zl = zapcore.DebugLevel
	default:
		zl = zapcore.DebugLevel - 1 // no Trace level in `zap`
	}
	return zl
}

func convertFieldToZap(field Field) zap.Field {
	k, v := field.Key, field.Value
	switch field.Type {
	case FieldTypeBool:
		return zap.Bool(k, v.(bool))
	case FieldTypeInt:
		return zap.Int(k, v.(int))
	case FieldTypeFloat:
		return zap.Float64(k, v.(float64))
	case FieldTypeStr:
		return zap.String(k, v.(string))
	case FieldTypeErr:
		return zap.NamedError(k, v.(error))
	case FieldTypeDur:
		return zap.Duration(k, v.(time.Duration))
	case FieldTypeTime:
		return zap.Time(k, v.(time.Time))
	default:
		return zap.Any(field.Key, field.Value)
	}
}

type zapLogger struct {
	logger *zap.Logger
}

func (l *zapLogger) log(lvl Level, msg string, fields ...Field) {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, convertFieldToZap(f))
	}
	l.logger.Log(convertLevelToZap(lvl), msg, zapFields...)
}

func (l *zapLogger) clone() loggerImpl {
	return &zapLogger{
		logger: l.logger,
	}
}

func (l *zapLogger) with(fields ...Field) loggerImpl {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, convertFieldToZap(f))
	}
	return &zapLogger{
		logger: l.logger.With(zapFields...),
	}
}

func (l *zapLogger) sync() error {
	return l.logger.Sync()
}
