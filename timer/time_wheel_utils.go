package timer

import (
	"time"
)

type ITimeTestHandler interface {
	Now() time.Time
	Get10Ms() time.Duration
}

var _timeTestHandler ITimeTestHandler = nil

func SetTimeTestHandler(handler ITimeTestHandler) {
	_timeTestHandler = handler
	get10Ms = _timeTestHandler.Get10Ms
}

var get10Ms = _get10Ms

func _get10Ms() time.Duration {
	return time.Duration(time.Now().UnixNano() / int64(time.Millisecond) / 10)
}
