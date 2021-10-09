package task

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/util/promise"
	"github.com/nomos/go-lokas/util"
)

type InputParam struct {
	Type     INPUT_TYPE
	Name     string
	Optional bool
}

type InputConfig struct {
	Inputs []InputParam
}

var _ lokas.ITaskPipeLine = (*PipeLine)(nil)

//任务基础类
type PipeLine struct {
	name         string
	idx          int
	context      lokas.IContext
	input        lokas.IContext
	output       lokas.IContext
	children     []lokas.ITaskPipeLine
	parent       lokas.ITaskPipeLine
	Order        TASK_ORDER
	FailedByPass bool
	execFunc     func() (IContext error, err error)
}

func (this *PipeLine) SetExecFunc(f func() (IContext error, err error)) {
	this.execFunc = f
}

func (this *PipeLine) SetInput(context lokas.IContext) {
	this.input = context
}

func (this *PipeLine) GetInput() lokas.IContext {
	return this.input
}

func (this *PipeLine) SetContext(context lokas.IContext) {
	this.context = context
}

func (this *PipeLine) Name() string {
	return this.name
}

func (this *PipeLine) Execute() *promise.Promise {
	all := []*promise.Promise{}
	for _, c := range this.children {
		all = append(all, c.Execute())
	}
	if this.Order == WATERFALL {
		return promise.Each(all...)
	} else if this.Order == PARALLEL {
		return promise.All(all...)
	}
	return nil
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
	id := this.idx - 1
	if this.parent == nil {
		return nil
	}
	return this.parent.GetChildById(id)
}

func (this *PipeLine) GetNext() lokas.ITaskPipeLine {
	id := this.idx + 1
	if this.parent == nil {
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
	if idx < 0 || idx > len(this.children)-1 {
		return nil
	}
	return this.children[idx]
}

func (this *PipeLine) GetChildByName(s string) lokas.ITaskPipeLine {
	for _, t := range this.children {
		if t.Name() == s {
			return t
		}
	}
	return nil
}

func (this *PipeLine) Add(flow lokas.ITaskPipeLine) lokas.ITaskPipeLine {
	flow.SetContext(this.context)
	this.children = append(this.children, flow)
	flow.SetIdx(len(this.children) - 1)
	return flow
}

func (this *PipeLine) Insert(flow lokas.ITaskPipeLine, idx int) lokas.ITaskPipeLine {
	if idx > len(this.children) {
		return flow
	}
	flow.SetContext(this.context)
	newArr := append([]lokas.ITaskPipeLine{}, this.children[:idx]...)
	newArr = append(newArr, flow)
	for _, v := range this.children[idx:] {
		v.SetIdx(v.Idx() + 1)
	}
	flow.SetIdx(idx)
	newArr = append(newArr, this.children[idx:]...)
	this.children = newArr
	return flow
}

func (this *PipeLine) RemoveAt(idx int) lokas.ITaskPipeLine {
	if idx > len(this.children)-1 {
		return nil
	}
	newArr := append([]lokas.ITaskPipeLine{}, this.children[:idx]...)
	for _, v := range this.children[idx+1:] {
		v.SetIdx(v.Idx() - 1)
	}
	newArr = append(this.children[idx+1:])
	ret := this.children[idx]
	ret.SetIdx(0)
	this.children = newArr
	return ret
}

func (this *PipeLine) Remove(flow lokas.ITaskPipeLine) lokas.ITaskPipeLine {
	for idx, v := range this.children {
		if flow == v {
			this.RemoveAt(idx)
			return flow
		}
	}
	return flow
}
