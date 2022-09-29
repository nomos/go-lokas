package log

import (
	"io"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap/zapcore"
)

type FileRotateOpt struct {
	MaxDay       int           //最多保留天数
	RotationSize int64         //最大尺寸
	RotationTime time.Duration //切割时间
	FileName     string        //文件名
	Json         bool          //是否用Json格式输出
	Format       string        //时间格式
	Filter       []func(ent zapcore.Entry) bool
}

type FileRotateHook struct {
	opt    *FileRotateOpt
	writer io.Writer
}

func NewFileRotationHook(opt *FileRotateOpt) *FileRotateHook {
	ret := &FileRotateHook{
		opt: opt,
	}
	ret.writer = getRotatelogsHook(opt.FileName, opt.MaxDay, opt.RotationTime, opt.RotationSize, opt.Format)
	return ret
}

func (this *FileRotateHook) WriteConsole(ent zapcore.Entry, p []byte) error {
	if this.opt.Json {
		return nil
	}
	_, err := this.writer.Write(p)
	return err
}

func (this *FileRotateHook) WriteJson(ent zapcore.Entry, p []byte) error {
	if this.opt.Json {
		for _, f := range this.opt.Filter {
			if f != nil && !f(ent) {
				return nil
			}
		}

		_, err := this.writer.Write(p)
		return err
	}
	return nil
}

func (this *FileRotateHook) WriteObject(ent zapcore.Entry, m map[string]interface{}) error {
	return nil
}

func getRotatelogsHook(filename string, maxDay int, rotationTime time.Duration, rotationSize int64, format string) io.Writer {
	if maxDay == 0 {
		maxDay = 7
	}
	if rotationTime == 0 {
		rotationTime = time.Hour
	}
	if rotationSize == 0 {
		rotationSize = 1024 * 1024
	}
	if format == "" {
		format = ".%Y-%m-%d_%H:%M:%S"
	}
	// 生成rotatelogs的Logger 实际生成的文件名 demo.log.YYmmddHH
	// demo.log是指向最新日志的链接
	// 保存7天内的日志，每1小时(整点)分割一次日志
	hook, err := rotatelogs.New(
		filename+format,
		rotatelogs.WithLinkName(filename),
		//最多保留多久
		rotatelogs.WithMaxAge(time.Hour*time.Duration(maxDay*24)),
		//多久做一次归档
		rotatelogs.WithRotationTime(rotationTime),
		rotatelogs.WithRotationSize(rotationSize),
	)

	if err != nil {
		panic(err)
	}
	return hook
}
