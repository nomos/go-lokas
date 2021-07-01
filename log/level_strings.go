package log

import (
	"github.com/nomos/go-lokas/log/color"
	"go.uber.org/zap/zapcore"
)

var (
	_levelToColor = map[zapcore.Level]color.Color{
		zapcore.DebugLevel:  color.Magenta,
		zapcore.InfoLevel:   color.Blue,
		zapcore.WarnLevel:   color.Yellow,
		zapcore.ErrorLevel:  color.Red,
		zapcore.DPanicLevel: color.Red,
		zapcore.PanicLevel:  color.Red,
		zapcore.FatalLevel:  color.Red,
	}
	_unknownLevelColor = color.Red

	_levelToLowercaseColorString = make(map[zapcore.Level]string, len(_levelToColor))
	_levelToCapitalColorString   = make(map[zapcore.Level]string, len(_levelToColor))
)

func init() {
	for level, color := range _levelToColor {
		_levelToLowercaseColorString[level] = color.Add(level.String())
		_levelToCapitalColorString[level] = color.Add(level.CapitalString())
	}
}
