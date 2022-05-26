package util

import (
	"fmt"
	"github.com/satori/go.uuid"
	"sort"
	"time"
)

type TimerState int
type TaskID int

const (
	TIMER_STOP   TimerState = 0
	TIMER_START  TimerState = 1
	TIMER_ONSTOP TimerState = 2
)

type TimerType int

const (
	TIMER_SYNC  TimerType = 0
	TIMER_ASYNC TimerType = 1
	TIMER_FIXED TimerType = 2
)

type ScheduleTask struct {
	name           string
	interval       int64
	activeTime     int64
	startTime      int64
	lastActiveTime int64
	task           func(activeTime int64, dt int64)
	count          int
	id             TaskID
}

type Timer struct {
	_ticker         *time.Ticker
	_timeScale      float32
	_updateTime     int64
	_startTime      int64
	_timeOffset     int64
	_runningTime    int64
	_lastStopTime   int64
	_lastUpdateTime int64
	_prevInterval   int64
	OnUpdate        func(dt int64, now int64)
	OnLateUpdate    func(dt int64, now int64)
	OnStart         func(now int64)
	OnStop          func(now int64)
	OnPause         func(now int64)
	OnResume        func(now int64)
	OnDestroy       func(now int64)
	_state          TimerState
	_type           TimerType
	_scheduleTasks  map[TaskID]*ScheduleTask //map[int]*ScheduleTask
	_ticks          int
	_taskIdGen      TaskID
	_sign           chan<- int
}

func (t *Timer) GetTimeScale() float32 {
	return t._timeScale
}

func (t *Timer) GetUpdateTime() int64 {
	return t._updateTime
}

func (t *Timer) GetTimeOffset() int64 {
	return t._timeOffset
}

func (t *Timer) GetRunningTime() int64 {
	return t._runningTime
}

func (t *Timer) GetPrevInterval() int64 {
	return t._prevInterval
}

func CreateTimer(updateTime int64, timeScale float32, sign chan<- int) *Timer {
	startTime := time.Now().UnixNano() / time.Millisecond.Nanoseconds()
	return &Timer{
		_ticker:         nil,
		_updateTime:     updateTime,
		_timeScale:      timeScale,
		_startTime:      startTime,
		_timeOffset:     0,
		_runningTime:    0,
		_lastUpdateTime: 0,
		_prevInterval:   0,
		OnUpdate:        nil,
		OnLateUpdate:    nil,
		OnStart:         nil,
		OnStop:          nil,
		OnPause:         nil,
		OnResume:        nil,
		OnDestroy:       nil,
		_state:          TIMER_STOP,
		_type:           TIMER_SYNC,
		_scheduleTasks:  map[TaskID]*ScheduleTask{},
		_ticks:          0,
		_taskIdGen:      0,
		_sign:           sign,
	}
}

func (t *Timer) Now() int64 {
	return time.Now().UnixNano() / time.Millisecond.Nanoseconds()
}

func (t *Timer) setOffset(offset int64) {
	t._runningTime -= t._timeOffset
	t._runningTime += offset
	t._timeOffset = offset
}

func (t *Timer) tickerUpdate() {
	t._lastUpdateTime = t.Now()
	t._ticker = time.NewTicker(time.Duration(t._updateTime * time.Millisecond.Nanoseconds()))
Loop:
	for {
		select {
		case <-t._ticker.C:
			t.instantUpdate()
			if t._state == TIMER_ONSTOP {
				break Loop
			}
		}
	}
	fmt.Println("done")
	t._ticker.Stop()
	t._ticker = nil
	t._state = TIMER_STOP
}

func (t *Timer) update() {
	if t._state == TIMER_ONSTOP {
		t.Stop()
		return
	}
	t._lastUpdateTime = t.Now()
	if t._updateTime <= t._prevInterval {
		t.instantUpdate()
		t.update()
	} else {
		time.Sleep(time.Duration((t._updateTime - t._prevInterval) * time.Millisecond.Nanoseconds()))
		t.instantUpdate()
		t.update()
	}
}

func (t *Timer) Start() {
	if t._state != TIMER_STOP {
		fmt.Print("Timer is Exist")
		return
	}
	t._state = TIMER_START
	t.tickerUpdate()
}

func (t *Timer) Pause() {
	t._state = TIMER_ONSTOP
}

func (t *Timer) Resume() {
	t.Start()
}

func (t *Timer) Stop() {
	t._state = TIMER_ONSTOP
	t._sign <- 0
}

func (t *Timer) Reset() {

}

func (t *Timer) instantUpdate() {
	now := t.Now()
	lastTime := t._lastUpdateTime
	//计算时间差值
	interval := now - lastTime
	//计算时间膨胀率
	interval = int64(float32(interval) / t._timeScale)
	//更新定时器运行时间
	t._runningTime += interval
	t._prevInterval = interval
	t._lastUpdateTime = now
	t.activeSchedule(t._prevInterval, t._runningTime)
	if t.OnUpdate != nil {
		t.OnUpdate(t._prevInterval, t._runningTime)
	}
	if t.OnLateUpdate != nil {
		t.OnLateUpdate(t._prevInterval, t._runningTime)
	}
}

func (t *Timer) createTask(name string, interval int64, count int, task func(int64, int64), startTime int64) *ScheduleTask {
	return &ScheduleTask{
		name:           name,
		interval:       interval,
		activeTime:     0,
		startTime:      startTime,
		lastActiveTime: startTime,
		task:           task,
		count:          count,
		id:             t._taskIdGen,
	}
}

func (t *Timer) activeSchedule(interval int64, now int64) {
	taskQueue := make([]interface{}, 0)
	removeTask := make([]TaskID, 0)
	for _, task := range t._scheduleTasks {
		for {
			//退出条件
			//如果任务的时间间隔大于当前时间减去上一次任务的间隔
			//或者任务的计数小于等于0
			if task.interval >= now-task.lastActiveTime || task.count <= 0 {
				break
			}
			vTask := *task
			toActiveTask := &vTask
			toActiveTask.activeTime = task.lastActiveTime - task.interval
			taskQueue = append(taskQueue, toActiveTask)
			task.count--
			task.lastActiveTime += task.interval
			if task.count <= 0 {
				removeTask = append(removeTask, task.id)
				break
			}
		}
		sort.Slice(taskQueue, func(i, j int) bool {
			a := taskQueue[i].(*ScheduleTask)
			b := taskQueue[j].(*ScheduleTask)
			if a.activeTime == b.activeTime {
				return a.id < b.id
			}
			return a.activeTime < b.activeTime
		})
		for _, taskInterface := range taskQueue {
			task := taskInterface.(*ScheduleTask)
			task.task(task.interval, task.activeTime)
		}
		taskQueue = make([]interface{}, 0)
		for _, taskId := range removeTask {
			t.Unschedule(taskId)
		}
		removeTask = make([]TaskID, 0)
	}
}

func (t *Timer) getTaskIdByName(name string) *ScheduleTask {
	for _, task := range t._scheduleTasks {
		if task.name == name {
			return task
		}
	}
	return nil
}

func (t *Timer) Schedule(name string, interval int64, count int, task func(int64, int64), delay int64, startTime int64) {
	if startTime == 0 {
		startTime = t._runningTime + delay
		delay = 0
	}
	if name == "" {
		uuidv1 := uuid.NewV1()
		name = uuidv1.String()
	}
	t._taskIdGen++
	if t.getTaskIdByName(name) != nil {
		panic("exist task name")
	}
	newTask := t.createTask(name, interval, count, task, startTime)
	t._scheduleTasks[t._taskIdGen] = newTask
}

func (t *Timer) Unschedule(task interface{}) {
	switch task.(type) {
	case TaskID:
		RemoveMapWithCondition(t._scheduleTasks, func(key TaskID, elem *ScheduleTask) bool {
			return key == task
		})
	case string:
		RemoveMapWithCondition(t._scheduleTasks, func(key TaskID, elem *ScheduleTask) bool {
			taskA := elem
			return taskA.name == task.(string)
		})
	case *ScheduleTask:
		RemoveMapElement(t._scheduleTasks, task.(*ScheduleTask))
	}
}

func (t *Timer) onStart() {
	if t.OnStart != nil {
		t.OnStart(t._runningTime)
	}
}

func (t *Timer) onResume() {
	if t.OnResume != nil {
		t.OnResume(t._runningTime)
	}
}

func (t *Timer) onStop() {
	if t.OnStop != nil {
		t.OnStop(t._runningTime)
	}
}

func (t *Timer) onPause() {
	if t.OnPause != nil {
		t.OnPause(t._runningTime)
	}
}

func (t *Timer) onDestroy() {
	if t.OnDestroy != nil {
		t.OnDestroy(t._runningTime)
	}
}
