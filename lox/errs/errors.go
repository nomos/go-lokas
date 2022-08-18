package errs

import (
	"github.com/nomos/go-lokas/protocol"
)

var (
	//inner error
	ERR_TYPE_NOT_FOUND  = protocol.CreateError(-1, "type not found")
	ERR_RPC_TIMEOUT     = protocol.CreateError(-2, "rpc timeout")
	ERR_PACKAGE_FORMAT  = protocol.CreateError(-3, "wrong packet format")
	ERR_INTERNAL_ERROR  = protocol.CreateError(-4, "internal error")
	ERR_ACTOR_NOT_FOUND = protocol.CreateError(-101, "actor not found")

	//cs error
	ERR_INTERNAL_SERVER        = protocol.CreateError(901, "服务器繁忙")
	ERR_TOKEN_EXPIRED          = protocol.CreateError(1001, "Token已过期")
	ERR_TOKEN_VALIDATE         = protocol.CreateError(1002, "Token无效")
	ERR_ACC_NOT_FOUND          = protocol.CreateError(1003, "找不到账户")
	ERR_GAME_ACC_NOT_FOUND     = protocol.CreateError(1004, "找不到游戏账户")
	ERR_GAME_ACC_EXIST         = protocol.CreateError(1011, "游戏账户已存在")
	ERR_GAME_ACC_CREATE_FAILED = protocol.CreateError(1012, "创建游戏账户失败")
	ERR_PARAM_NOT_EXIST        = protocol.CreateError(1101, "参数不存在")
	ERR_PARAM_TYPE             = protocol.CreateError(1102, "参数错误")
	ERR_ACC_AUTH               = protocol.CreateError(1103, "账户不存在或者密码错误")
	ERR_PASSWORD_EMPTY         = protocol.CreateError(1104, "需要输入密码")
	ERR_AUTH_FAILED            = protocol.CreateError(1201, "验证失败")
	ERR_MSG_FORMAT             = protocol.CreateError(1202, "数据格式错误")
	ERR_PROTOCOL_NOT_FOUND     = protocol.CreateError(1203, "协议未找到")
	ERR_NAME_EXIST             = protocol.CreateError(1222, "名称重复")
	ERR_MULTI_DEV_LOG          = protocol.CreateError(1204, "multi device login")
	ERR_SERVICE_NOT_FIND       = protocol.CreateError(1205, "服务器未找到")
)
