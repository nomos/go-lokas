package protocol

import (
	"strconv"
)

type ErrCode int

var (
	ERR_SUCC = CreateError(0, "successful")

	//inner error
	ERR_TYPE_NOT_FOUND  = CreateError(-1, "type not found")
	ERR_RPC_TIMEOUT     = CreateError(-2, "rpc timeout")
	ERR_PACKAGE_FORMAT  = CreateError(-3, "wrong packet format")
	ERR_INTERNAL_ERROR  = CreateError(-4, "internal error")
	ERR_ACTOR_NOT_FOUND = CreateError(-101, "actor not found")
	ERR_RPC_FAILED      = CreateError(-102, "rpc failed")

	ERR_JSON_MARSHAL_FAILED = CreateError(-202, "json marshal failed")
	// msg
	ERR_MSG_ROUTE_NOT_FOUND = CreateError(-6001, "msg route local not found avatar")

	// service
	ERR_REGISTER_SERVICE_DUPLICATED   = CreateError(-7001, "service register duplicate")
	ERR_REGISTER_SERVICE_INFO_INVALID = CreateError(-7002, "service info invalid")
	ERR_REGISTER_SERVICE_NOT_FOUND    = CreateError(-7003, "service registered not found")

	ERR_REGISTER_ROUTE_USER_DUPLICATED = CreateError(-7101, "user route register duplicate")

	ERR_ETCD_ERROR   = CreateError(201, "数据错误")
	ERR_DB_ERROR     = CreateError(202, "数据库错误")
	ERR_CONFIG_ERROR = CreateError(203, "配置错误")

	ERR_MSG_LEN_INVALID = CreateError(301, "数据长度无效")

	//cs error
	ERR_INTERNAL_SERVER        = CreateError(901, "服务器繁忙")
	ERR_TOKEN_EXPIRED          = CreateError(1001, "Token已过期")
	ERR_TOKEN_VALIDATE         = CreateError(1002, "Token无效")
	ERR_ACC_NOT_FIND           = CreateError(1003, "找不到账户")
	ERR_GAME_ACC_NOT_FOUND     = CreateError(1004, "找不到游戏账户")
	ERR_GAME_ACC_EXIST         = CreateError(1011, "游戏账户已存在")
	ERR_GAME_ACC_CREATE_FAILED = CreateError(1012, "创建游戏账户失败")
	ERR_PARAM_NOT_EXIST        = CreateError(1101, "参数不存在")
	ERR_PARAM_TYPE             = CreateError(1102, "参数错误")
	ERR_ACC_AUTH               = CreateError(1103, "账户不存在或者密码错误")
	ERR_PASSWORD_EMPTY         = CreateError(1104, "需要输入密码")
	ERR_AUTH_FAILED            = CreateError(1201, "验证失败")
	ERR_MSG_FORMAT             = CreateError(1202, "数据格式错误")
	ERR_PROTOCOL_NOT_FOUND     = CreateError(1203, "协议未找到")
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
	Is(e error) bool
}

var _ ISerializable = &ErrMsg{}

type ErrMsg struct {
	Code    int32
	Message string
}

func (this *ErrMsg) ErrCode() int {
	return int(this.Code)
}

func (this *ErrMsg) Is(err error) bool {
	if e, ok := err.(IError); ok {
		if e.ErrCode() == this.ErrCode() {
			return true
		}
	}
	return false
}

func (this *ErrMsg) Unmarshal(from []byte) error {
	return Unmarshal(from, this)
}

func (this *ErrMsg) Marshal() ([]byte, error) {
	return MarshalBinary(this)
}

func (this *ErrMsg) GetId() (BINARY_TAG, error) {
	return TAG_Error, nil
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
