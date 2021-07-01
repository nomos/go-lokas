package log

import (
	"go.uber.org/zap/zapcore"
	"os"
)

const (
	ConsoleEncoder string = "console"
	JsonEncoder    string = "json"
	ObjEncoder     string = "obj"
)

type MongoConf struct {
	Enable     bool
	Addr       string
	Db         string
	Collection string
	Expires    string
	User       string
	Pwd        string
}

type FileRotateConf struct {
}

type LogConfig struct {
	Title string
	EncodeCfg      *zapcore.EncoderConfig
	Console        bool
	JsonConsole	   bool
	Json           bool
	Object         bool
	FileRotateConf FileRotateConf
}

type ZapCompose struct {
	cfg           *LogConfig
	consoleEnc    zapcore.Encoder
	composeEnc    *ComposeEncoder
	enableJson    bool
	enableConsole bool
	enableObj     bool
	hooks         []Hook
	stdOut        zapcore.WriteSyncer
}

var _ zapcore.Core = (*ZapCompose)(nil)

func NewZapCompose(cfg *LogConfig, hook ...Hook) *ZapCompose {
	ret := &ZapCompose{
		cfg:           cfg,
		composeEnc:    NewComposeEncoder(cfg.EncodeCfg, cfg.Json, cfg.Object),
		enableJson:    cfg.Json,
		enableConsole: cfg.Console,
		enableObj:     cfg.Object,
		hooks:         []Hook{},

		stdOut: zapcore.Lock(os.Stdout),
	}
	if cfg.JsonConsole {
		ret.consoleEnc =    zapcore.NewJSONEncoder(*cfg.EncodeCfg)
	} else {
		ret.consoleEnc =    zapcore.NewConsoleEncoder(*cfg.EncodeCfg)
	}
	for _, h := range hook {
		ret.hooks = append(ret.hooks, h)
	}
	return ret
}

func (this *ZapCompose) Enabled(level zapcore.Level) bool {
	return true
}

func (this *ZapCompose) clone() *ZapCompose {
	return &ZapCompose{
		consoleEnc:   this.consoleEnc.Clone(),
		composeEnc:   this.composeEnc.Clone(),
	}
}

func addFields(enc zapcore.ObjectEncoder, fields []zapcore.Field) {
	for i := range fields {
		fields[i].AddTo(enc)
	}
}

func (this *ZapCompose) With(fields []zapcore.Field) zapcore.Core {
	clone := this.clone()
	addFields(clone.consoleEnc, fields)
	return clone
}

func (this *ZapCompose) Check(ent zapcore.Entry, che *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if this.Enabled(ent.Level) {
		return che.AddCore(ent, this)
	}
	return che
}

func (this *ZapCompose) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	if this.enableConsole {
		buf, err := this.consoleEnc.EncodeEntry(ent, fields)
		if err != nil {
			return err
		}
		_,err = this.stdOut.Write(buf.Bytes())
		if err != nil {
			return err
		}
		for _, h := range this.hooks {
			err := h.WriteConsole(ent, []byte(ent.Message))
			if err != nil {
				return err
			}
		}
		buf.Free()
		if err != nil {
			return err
		}
	}

	if this.enableJson || this.enableObj {
		buf, err := this.composeEnc.EncodeEntry(ent, fields)
		if err != nil {
			return err
		}
		if this.enableJson {
			for _, h := range this.hooks {
				err := h.WriteJson(ent, buf.Bytes())
				if err != nil {
					return err
				}
			}
		}
		buf.Free()
		if this.enableObj {
			for _, h := range this.hooks {
				err := h.WriteObject(ent, this.composeEnc.Fields)
				if err != nil {
					return err
				}
			}
			m := make(map[string]interface{})
			this.composeEnc.Fields = m
			this.composeEnc.cur = m
		}

	}
	if ent.Level > zapcore.ErrorLevel {
		// Since we may be crashing the program, sync the output. Ignore Sync
		// errors, pending a clean solution to issue #370.
		this.Sync()
	}
	return nil
}

func (this *ZapCompose) Sync() error {
	return nil
}
