package task

import "github.com/nomos/go-lokas/protocol"

type TASK_ORDER protocol.Enum

const (
	PARALLEL  TASK_ORDER = iota //并行
	WATERFALL                   //串行
)

type INPUT_TYPE protocol.Enum

const (
	FROM_CONTEXT      INPUT_TYPE = iota //从环境变量
	FROM_PARENT_INPUT                   //从父任务的Input
	FROM_PREV                           //从上一个任务
	FROM_SIBLING                        //从同级任务
)
