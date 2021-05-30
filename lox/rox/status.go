package rox

import "net/http"

func text(w http.ResponseWriter, code int, msg string) {
	writer := w.(ResponseWriter)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	writer.SetText(msg)
}

func write(w http.ResponseWriter, code int, msg []interface{}) {
	writer := w.(ResponseWriter)
	if len(msg) == 0 {
		text(w, code, http.StatusText(code))
		return
	}

	switch msg[0].(type) {
	case string:
		text(w, code, msg[0].(string))
	default:
		JSON(writer, msg[0], code)
	}
}

func Continue(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusContinue, msg)
}

func SwitchingProtocols(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusSwitchingProtocols, msg)
}

func OK(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusOK, msg)
}

func Created(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusCreated, msg)
}

func Accepted(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusAccepted, msg)
}

func NonAuthoritativeInfo(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusNonAuthoritativeInfo, msg)
}

func NoContent(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusNoContent, msg)
}

func ResetContent(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusResetContent, msg)
}

func PartialContent(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusPartialContent, msg)
}

func MultipleChoices(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusMultipleChoices, msg)
}

func MovedPermanently(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusMovedPermanently, msg)
}

func Found(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusFound, msg)
}

func SeeOther(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusSeeOther, msg)
}

func NotModified(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusNotModified, msg)
}

func UseProxy(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusUseProxy, msg)
}

func TemporaryRedirect(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusTemporaryRedirect, msg)
}

func BadRequest(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusBadRequest, msg)
}

func Unauthorized(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusUnauthorized, msg)
}

func PaymentRequired(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusPaymentRequired, msg)
}

func Forbidden(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusForbidden, msg)
}

func NotFound(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusNotFound, msg)
}

func MethodNotAllowed(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusMethodNotAllowed, msg)
}

func NotAcceptable(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusNotAcceptable, msg)
}

func ProxyAuthRequired(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusProxyAuthRequired, msg)
}

func RequestTimeout(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusRequestTimeout, msg)
}

func Conflict(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusConflict, msg)
}

func Gone(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusGone, msg)
}

func LengthRequired(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusLengthRequired, msg)
}

func PreconditionFailed(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusPreconditionFailed, msg)
}

func RequestEntityTooLarge(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusRequestEntityTooLarge, msg)
}

func RequestURITooLong(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusRequestURITooLong, msg)
}

func UnsupportedMediaType(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusUnsupportedMediaType, msg)
}

func RequestedRangeNotSatisfiable(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusRequestedRangeNotSatisfiable, msg)
}

func ExpectationFailed(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusExpectationFailed, msg)
}

func Teapot(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusTeapot, msg)
}

func InternalServerError(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusInternalServerError, msg)
}

func NotImplemented(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusNotImplemented, msg)
}

func BadGateway(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusBadGateway, msg)
}

func ServiceUnavailable(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusServiceUnavailable, msg)
}

func GatewayTimeout(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusGatewayTimeout, msg)
}

func HTTPVersionNotSupported(w http.ResponseWriter, msg ...interface{}) {
	write(w, http.StatusHTTPVersionNotSupported, msg)
}
