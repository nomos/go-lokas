package log

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"reflect"
	"regexp"
	"sync"
	"time"
)

var __hooks []Hook = make([]Hook, 0)
var once sync.Once
var _logger *ComposeLogger
var _console bool

type titleEncoder struct {
	title                       string
	levelToLowercaseColorString map[zapcore.Level]string
	levelToCapitalColorString   map[zapcore.Level]string
}

func newTitleEncoder(title string) *titleEncoder {
	ret := &titleEncoder{
		title:                       "",
		levelToLowercaseColorString: make(map[zapcore.Level]string, len(_levelToColor)),
		levelToCapitalColorString:   make(map[zapcore.Level]string, len(_levelToColor)),
	}
	for level, color := range _levelToColor {
		ret.levelToLowercaseColorString[level] = color.Add("[" + level.String() + "]" + title)
		ret.levelToCapitalColorString[level] = color.Add("[" + level.CapitalString() + "]" + title)
	}
	return ret
}

func (this *titleEncoder) Encode(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	s, ok := this.levelToCapitalColorString[l]
	if !ok {
		s = _unknownLevelColor.Add(l.CapitalString() + this.title)
	}
	enc.AppendString(s)
}

func DefaultConsoleConfig(title string) *zapcore.EncoderConfig {
	return &zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    newTitleEncoder(title).Encode, //这里可以指定颜色
		EncodeTime:     TimeEncoder,                   // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 全路径编码器
	}
}

var DefaultJsonConfig = zapcore.EncoderConfig{
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.CapitalLevelEncoder, //这里可以指定颜色
	EncodeTime:     nil,                         // ISO8601 UTC 时间格式
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder, // 全路径编码器
}

var DefaultJsonConfigWithTime = zapcore.EncoderConfig{
	TimeKey:        "time",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.CapitalLevelEncoder, //这里可以指定颜色
	EncodeTime:     TimeEncoder,                 // ISO8601 UTC 时间格式
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder, // 全路径编码器
}

func NewEncoderConfig(title string, console bool) *zapcore.EncoderConfig {
	if console {
		return DefaultConsoleConfig(title)
	} else {
		return &DefaultJsonConfig
	}
}

func CustomConfig(json, console, jsonConsole, object bool, config *zapcore.EncoderConfig) *LogConfig {
	return &LogConfig{
		Title:       "",
		EncodeCfg:   config,
		Console:     console,
		JsonConsole: jsonConsole,
		Json:        json,
		Object:      object,
	}
}

func ConsoleConfig(title string) *LogConfig {
	return &LogConfig{
		Title:       "",
		EncodeCfg:   NewEncoderConfig(title, true),
		Console:     true,
		JsonConsole: false,
		Json:        false,
		Object:      false,
	}
}

func DefaultConfig(title string) *LogConfig {
	return &LogConfig{
		Title:       "",
		EncodeCfg:   NewEncoderConfig(title, false),
		Console:     true,
		JsonConsole: true,
		Json:        true,
		Object:      false,
	}
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func newLogConfig() zap.Config {
	return zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       true,
		DisableCaller:     false,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "console",                   // 输出格式 console 或 json
		EncoderConfig:     *NewEncoderConfig("", true), // 编码器配置
		InitialFields:     map[string]interface{}{},    // 初始化字段，如：添加一个服务器名称
		OutputPaths:       []string{"stdout"},          // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths:  []string{"stderr"},
	}
}

//var exportConfig = zap.Config{
//	Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
//	Development:       true,
//	DisableCaller:     false,
//	DisableStacktrace: true,
//	Sampling:          nil,
//	Encoding:         "json",                // 输出格式 console 或 json
//	EncoderConfig:    encoderConfig,            // 编码器配置
//	InitialFields:    map[string]interface{}{}, // 初始化字段，如：添加一个服务器名称
//}

type Field struct {
	Key   string
	Value interface{}
}

type Fields map[string]interface{}

func InitDefault(dev bool, console bool, title string, hooks ...Hook) {
	for _, h := range __hooks {
		hooks = append(hooks, h)
	}
	if console {
		_console = true
		_logger = NewComposeLogger(dev, ConsoleConfig(title), 2, hooks...)
	} else {
		_console = false
		_logger = NewComposeLogger(dev, DefaultConfig(title), 2, hooks...)
	}
}

func IsConsole() bool {
	return _console
}

func Init(dev bool, conf *LogConfig, hooks ...Hook) {
	for _, h := range __hooks {
		hooks = append(hooks, h)
	}
	_logger = NewComposeLogger(dev, conf, 2, hooks...)
}

func AddHook(hook Hook) {
	if _logger == nil {
		Init(true, DefaultConfig(""))
	}
	for _, v := range __hooks {
		if v == hook {
			return
		}
	}
	__hooks = append(__hooks, hook)
	_logger.AddHook(hook)
}

func RemoveHook(hook Hook) {
	if _logger == nil {
		Init(true, DefaultConfig(""))
	}
	ret := make([]Hook, 0)
	for index, v := range __hooks {
		ret = __hooks[:index]
		if v == hook {
			if index == len(__hooks)-2 {
				break
			}
			for _, w := range __hooks[index+1:] {
				ret = append(ret, w)
			}
			break
		}
	}
	_logger.RemoveHook(hook)
}

func Close() {
	if _logger != nil {
	}
}

func SetDefaultLogger(logger *ComposeLogger) {
	_logger = logger
}

func DefaultLogger() *ComposeLogger {
	if _logger == nil {
		Init(true, DefaultConfig(""))
	}
	return _logger
}

func WithFields(fields Fields) *ComposeLogger {
	return _logger.WithFields(fields)
}

func WithField(key string, value interface{}) *ComposeLogger {
	return _logger.WithField(key, value)
}

var reg1 = regexp.MustCompile(`(%v)|(%[+]v)|(%[#]v )|(%T)|(%t)|(%b )|(%c)|(%d)|(%o)|(%q)|(%x)|(%X)|(%U)|(%b)|(%e)|(%E)|(%[[.]|(0-9)]*f)|(%[[.]|(0-9)]*g)|(%[[.]|(0-9)]*G)|(%s)|(%q)|(%f)|(%x)|(%X)|(%p)`)

func sprintf(args []interface{}) interface{} {
	ret := make([]interface{}, 0)
	if len(args) == 1 {
		return args[0]
	}
	if len(args) > 1 {
		if reflect.TypeOf(args[0]).Kind() == reflect.String {
			argstr := args[0].(string)
			if reg1.MatchString(argstr) {
				return fmt.Sprintf(argstr, args[1:]...)
			}
		}
	}
	for _, arg := range args {
		text, err := json.Marshal(arg)
		if err != nil {
			ret = append(ret, arg)
		}
		ret = append(ret, string(text))
	}
	return ret
}

type LoggerWithFields struct {
	fields []zap.Field
}

func (this *LoggerWithFields) reset() {
	this.fields = []zap.Field{}
}

type TestLogger struct {
	_logger *zap.Logger
	enabled bool
}

type ILogger interface {
	Clear()
	Infof(args ...interface{})
	Debugf(args ...interface{})
	Warnf(args ...interface{})
	Errorf(args ...interface{})
	Panicf(args ...interface{})
	Fatalf(args ...interface{})
	Info(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field) error
	Panic(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	WriteString(s string)
	Write(p []byte) (int, error)
}

type StdLogger interface {
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Print(v ...interface{})
	Printf(format string, v ...interface{})
}

var _ StdLogger = (*TestLogger)(nil)

func NewAstilecTronLogger(enabled bool) *TestLogger {
	ret := &TestLogger{enabled: enabled}
	ret._logger, _ = newLogConfig().Build(zap.AddCallerSkip(3))
	return ret
}

func (this *TestLogger) Fatal(fields ...interface{}) {
	if !this.enabled {
		return
	}
	this._logger.Fatal(fields[0].(string))
}

func (this *TestLogger) Fatalf(msg string, fields ...interface{}) {
	if !this.enabled {
		return
	}
	this._logger.Fatal(fmt.Sprintf(msg, fields))
}

func (this *TestLogger) Print(fields ...interface{}) {
	if !this.enabled {
		return
	}
	this._logger.Info(fields[0].(string))
}

func (this *TestLogger) Printf(msg string, fields ...interface{}) {
	if !this.enabled {
		return
	}
	this._logger.Info(fmt.Sprintf(msg, fields))
}

func Infof(args ...interface{}) {
	if _logger == nil {
		Init(true, DefaultConfig(""))
	}
	_logger.Infof(args...)
}

func Warnf(args ...interface{}) {
	if _logger == nil {
		Init(true, DefaultConfig(""))
	}
	_logger.Warnf(args...)
}

func Debugf(args ...interface{}) {
	if _logger == nil {
		Init(true, DefaultConfig(""))
	}
	_logger.Debugf(args...)
}

func Errorf(args ...interface{}) {
	if _logger == nil {
		Init(true, DefaultConfig(""))
	}
	_logger.Errorf(args...)
}

func Panicf(args ...interface{}) {
	if _logger == nil {
		Init(true, DefaultConfig(""))
	}
	_logger.Panicf(args...)
}

func Fatalf(args ...interface{}) {
	if _logger == nil {
		Init(true, DefaultConfig(""))
	}
	_logger.Fatalf(args...)
}

func Warn(msg string, fields ...zap.Field) {
	if _logger == nil {
		Init(true, DefaultConfig(""))
	}
	_logger.Warn(msg, fields...)
	//_exportLogger.Warn(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	if _logger == nil {
		Init(true, DefaultConfig(""))
	}
	_logger.Info(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	if _logger == nil {
		Init(true, DefaultConfig(""))
	}
	_logger.Debug(msg, fields...)
}

func Error(msg string, fields ...zap.Field) error {
	if _logger == nil {
		Init(true, DefaultConfig(""))
	}
	return _logger.Error(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	if _logger == nil {
		Init(true, DefaultConfig(""))
	}
	_logger.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	if _logger == nil {
		Init(true, DefaultConfig(""))
	}
	_logger.Fatal(msg, fields...)
}

func Sync() error {
	if _logger == nil {
		return nil
	}
	return _logger.Sync()
}

func PrettyStruct(s interface{}) string {
	return fmt.Sprintf("%+v", s)
}
