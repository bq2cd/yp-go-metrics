package log

import (
	"io"
	"maps"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	zerolog.TimestampFieldName = "ts"
	zerolog.MessageFieldName = "msg"
	zerolog.CallerSkipFrameCount += 2 // 1 from loggerImpl.log, 1 from EventBuilder.Msg/Send
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(filepath.Dir(file)) + "/" + filepath.Base(file) + ":" + strconv.Itoa(line)
	}
}

// NewZeroLogger creates a logger backed by `zerolog.Logger`
func NewZeroLogger(w io.Writer) (Logger, error) {
	logger := zerolog.New(w)
	l := &baseLogger{impl: &zeroLogger{
		logger:      logger,
		fieldsByKey: make(map[string]Field, 8), // zerolog does not have fields deduplication
	}}
	return l, nil
}

func convertLevelToZero(lvl Level) zerolog.Level {
	var zl zerolog.Level
	switch lvl {
	case LevelFatal:
		zl = zerolog.FatalLevel
	case LevelError:
		zl = zerolog.ErrorLevel
	case LevelInfo:
		zl = zerolog.InfoLevel
	case LevelDebug:
		zl = zerolog.DebugLevel
	default:
		zl = zerolog.TraceLevel
	}
	return zl
}

func convertFieldToZeroEvent(evt *zerolog.Event, field Field) *zerolog.Event {
	k, v := field.Key, field.Value
	switch field.Type {
	case FieldTypeBool:
		return evt.Bool(k, v.(bool))
	case FieldTypeInt:
		return evt.Int(k, v.(int))
	case FieldTypeFloat:
		return evt.Float64(k, v.(float64))
	case FieldTypeStr:
		return evt.Str(k, v.(string))
	case FieldTypeErr:
		return evt.Err(v.(error))
	case FieldTypeDur:
		return evt.Dur(k, v.(time.Duration))
	case FieldTypeTime:
		return evt.Time(k, v.(time.Time))
	default:
		return evt.Any(field.Key, field.Value)
	}
}

type zeroLogger struct {
	logger      zerolog.Logger
	fieldsByKey map[string]Field
}

func (l *zeroLogger) addFields(fields ...Field) {
	for _, f := range fields {
		l.fieldsByKey[f.Key] = f
	}
}

func (l *zeroLogger) log(lvl Level, msg string, fields ...Field) {
	l.addFields(fields...)
	zlvl := convertLevelToZero(lvl)
	evt := l.logger.WithLevel(zlvl).Timestamp().Caller()
	for _, f := range l.fieldsByKey {
		evt = convertFieldToZeroEvent(evt, f)
	}
	evt.Msg(msg)
}

func (l *zeroLogger) clone() loggerImpl {
	newLogger := &zeroLogger{
		logger:      l.logger.With().Logger(),
		fieldsByKey: make(map[string]Field, len(l.fieldsByKey)*2),
	}
	maps.Copy(newLogger.fieldsByKey, l.fieldsByKey)
	return newLogger
}

func (l *zeroLogger) with(fields ...Field) loggerImpl {
	newLogger := l.clone().(*zeroLogger)
	newLogger.addFields(fields...)
	return newLogger
}

func (l *zeroLogger) sync() {
	// not supported
}
