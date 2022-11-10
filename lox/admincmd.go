package lox

import (
	"reflect"
	"strconv"

	"github.com/nomos/go-lokas/cmds"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
)

type AdminCommand struct {
	Command  string
	Username string
	Params   []string
}

func (this *AdminCommand) GetId() (protocol.BINARY_TAG, error) {
	return TAG_ADMIN_CMD, nil
}

func (this *AdminCommand) Serializable() protocol.ISerializable {
	return this
}

func NewAdminCommand(cmd string, username string, params ...interface{}) *AdminCommand {
	pArr := []string{}
	for _, p := range params {
		switch reflect.TypeOf(p).Kind() {
		case reflect.String:
			pArr = append(pArr, p.(string))
		case reflect.Int:
			pArr = append(pArr, strconv.Itoa(p.(int)))
		case reflect.Int32:
			pArr = append(pArr, strconv.Itoa(int(p.(int32))))
		case reflect.Float64:
			pArr = append(pArr, strconv.FormatFloat(p.(float64), 'f', 10, 64))
		case reflect.Bool:
			if p.(bool) {
				pArr = append(pArr, "true")
			} else {
				pArr = append(pArr, "false")
			}
		default:
			log.Panic("unsupported param type:" + reflect.TypeOf(p).Kind().String())
		}

	}
	return &AdminCommand{
		Command:  cmd,
		Username: username,
		Params:   pArr,
	}
}

type AdminCommandResult struct {
	Command  string
	Username string
	Success  bool
	Data     []byte
}

func NewAdminCommandResult(cmd *AdminCommand, success bool, data []byte) *AdminCommandResult {
	ret := &AdminCommandResult{
		Command:  cmd.Command,
		Username: cmd.Username,
		Success:  success,
		Data:     data,
	}

	return ret
}

func (this *AdminCommandResult) GetId() (protocol.BINARY_TAG, error) {
	return TAG_ADMIN_CMD_RESULT, nil
}

func (this *AdminCommandResult) Serializable() protocol.ISerializable {
	return this
}

func (this *AdminCommand) ParamsValue() *cmds.ParamsValue {
	params := []interface{}{}
	for _, v := range this.Params {
		params = append(params, v)
	}
	return cmds.NewParamsValue(this.Command, params...)
}
