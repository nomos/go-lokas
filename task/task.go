package task

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/util"
)

type Input struct {
	Type INPUT_TYPE
	Name string
	Value interface{}
}

var _ lokas.ITaskPipeLine = (*PipeLine)(nil)

func (this *PipeLine) Name() string {
	return this.name
}

//任务基础类
type PipeLine struct {
	name         string
	idx int
	context      lokas.IContext
	children     []lokas.ITaskPipeLine
	parent       lokas.ITaskPipeLine
	Order        TASK_ORDER
	FailedByPass bool
}

func (this *PipeLine) Idx() int {
	return this.idx
}

func (this *PipeLine) SetIdx(i int) {
	this.idx = i
}

func (this *PipeLine) SetName(s string) {
	this.name = s
}

func (this *PipeLine) GetContext() lokas.IContext {
	return this.context
}

func (this *PipeLine) GetParent() lokas.ITaskPipeLine {
	return this.parent
}

func (this *PipeLine) GetPrev() lokas.ITaskPipeLine {
	id:=this.idx-1
	if this.parent==nil {
		return nil
	}
	return this.parent.GetChildById(id)
}

func (this *PipeLine) GetNext() lokas.ITaskPipeLine {
	id:=this.idx+1
	if this.parent==nil {
		return nil
	}
	return this.parent.GetChildById(id)
}

func (this *PipeLine) GetSibling(idx int) lokas.ITaskPipeLine {
	if util.IsNil(this.parent) {
		return nil
	}
	return this.parent.GetChildById(idx)
}

func (this *PipeLine) GetChildren() []lokas.ITaskPipeLine {
	return this.children
}

func (this *PipeLine) GetChildById(idx int) lokas.ITaskPipeLine {
	if idx<0||idx>len(this.children)-1 {
		return nil
	}
	return this.children[idx]
}

func (this *PipeLine) GetChildByName(s string) lokas.ITaskPipeLine {
	for _,t:=range this.children {
		if t.Name()==s {
			return t
		}
	}
	return nil
}

func (this *PipeLine) Add(flow lokas.ITaskPipeLine) {
	this.children = append(this.children, flow)
	flow.SetIdx(len(this.children)-1)
}

func (this *PipeLine) Insert(flow lokas.ITaskPipeLine, idx int) {

}

func (this *PipeLine) Remove(flow lokas.ITaskPipeLine) {

}
