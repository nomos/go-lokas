package protocol

import (
	"strconv"
)

type ErrCode int

var (
	//inner error
	ErrTypeNotFound      = CreateError(-1, "type not found")
	ErrRpcTimeOut        = CreateError(-2, "rpc timeout")
	ErrPacketWrongFormat = CreateError(-3, "wrong packet format")
	ErrInternalError     = CreateError(-4, "internal error")
	ErrActorNotExist     = CreateError(-101, "actor not found")

	//cs error
	ErrInternalServer       = CreateError(901, "服务器繁忙")
	ErrTokenExpired        = CreateError(1001, "Token已过期")
	ErrTokenValidate       = CreateError(1002, "Token无效")
	ErrAccNotFind          = CreateError(1003, "找不到账户")
	ErrGameAccNotFind      = CreateError(1004, "找不到游戏账户")
	ErrGameAccAlreadyExist = CreateError(1011, "游戏账户已存在")
	ErrGameAccCreateFailed = CreateError(1012, "创建游戏账户失败")
	ErrParamNotExist       = CreateError(1101, "参数不存在")
	ErrParamError          = CreateError(1102, "参数错误")
	ErrPassAuthError       = CreateError(1103, "账户不存在或者密码错误")
	ErrPasswordNeeded      = CreateError(1104, "需要输入密码")
	ErrAuthFailed          = CreateError(1201, "验证失败")
	ErrMsgFormat           = CreateError(1202, "数据格式错误")
)

func (this ErrCode) Error() string {
	return strconv.Itoa(int(this)) + ":" + predefined_errors[this]
}

func (this ErrCode) ErrCode() int {
	return int(this)
}

func (this ErrCode) NewErrMsg() *ErrMsg {
	return &ErrMsg{
		Code:    int32(this),
		Message: this.Error(),
	}
}

func (this ErrCode) Is(err error) bool {
	if e, ok := err.(IError); ok {
		if e.ErrCode() == this.ErrCode() {
			return true
		}
	}
	return false
}

var predefined_errors = map[ErrCode]string{}

type IError interface {
	Error() string
	ErrCode() int
}

var _ ISerializable = &ErrMsg{}

type ErrMsg struct {
	Code    int32
	Message string
}

func (this *ErrMsg) ErrCode() int {
	return int(this.Code)
}

func (this *ErrMsg) Unmarshal(from []byte) error {
	return Unmarshal(from, this)
}

func (this *ErrMsg) Marshal() ([]byte, error) {
	return MarshalBinary(this)
}

func (this *ErrMsg) GetId() (BINARY_TAG, error) {
	return GetCmdIdFromType(this)
}

func (this *ErrMsg) Serializable() ISerializable {
	return this
}

func NewError(code ErrCode) *ErrMsg {
	return &ErrMsg{
		Code:    int32(code),
		Message: code.Error(),
	}
}

func NewErrorMsg(code int32, message string) *ErrMsg {
	return &ErrMsg{
		Code:    code,
		Message: message,
	}
}

func (this *ErrMsg) Error() string {
	return this.Message
}

func CreateError(code ErrCode, msg string) ErrCode {
	if predefined_errors[code] != "" && predefined_errors[code] != msg {
		panic("error conflict, errcode:" + strconv.Itoa(int(code)) + " msg:" + msg)
	}
	predefined_errors[code] = msg
	return code
}
