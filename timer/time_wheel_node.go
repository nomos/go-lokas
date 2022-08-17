package timer

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/go-lokas/util/bitset"
	"github.com/nomos/go-lokas/util/xmath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	haveStop = uint32(1)
)

// 先使用sync.Mutex实现功能
// 后面使用cas优化
type Time struct {
	timeNode
	sync.Mutex

	// |---16bit---|---16bit---|------32bit-----|
	// |---level---|---index---|-------seq------|
	// level 在near盘子里就是1, 在T2ToTt[0]盘子里就是2起步
	// index 就是各自盘子的索引值
	// seq   自增id
	version uint64
}

func newTimeHead(level uint64, index uint64) *Time {
	head := &Time{}
	head.version = genVersionHeight(level, index)
	head.Init()
	return head
}

func genVersionHeight(level uint64, index uint64) uint64 {
	return level<<(32+16) | index<<32
}

func (t *Time) lockPushBack(node *timeNode, level uint64, index uint64) {
	t.Lock()
	defer t.Unlock()
	if atomic.LoadUint32(&node.stop) == haveStop {
		return
	}

	t.AddTail(&node.Head)
	atomic.StorePointer(&node.list, unsafe.Pointer(t))
	//更新节点的version信息
	atomic.StoreUint64(&node.version, atomic.LoadUint64(&t.version))
}

var _ TimeNoder = (*timeNode)(nil)

const ALL_WEEK bitset.BitSet = 0b1111111
const ALL_MONTH bitset.BitSet = 0b111111111111
const ALL_DAY bitset.BitSet = 0b1111111111111111111111111111111
const ALL_HOUR bitset.BitSet = 0b111111111111
const ALL_MINUTE bitset.BitSet = 0b111111111111111111111111111111111111111111111111111111111111
const ALL_SECOND bitset.BitSet = ALL_MINUTE

type timeNode struct {
	expire uint64
	// userExpire time.Duration
	callback func(TimeNoder)
	stop     uint32
	list     unsafe.Pointer //存放表头信息
	version  uint64         //保存节点版本信息
	// isSchedule bool
	delay        uint64
	interval     uint64
	loopCur      uint64
	loopMax      uint64
	isCron       bool
	month        bitset.BitSet
	weekday      bitset.BitSet
	monthday     bitset.BitSet
	hour         bitset.BitSet
	minute       bitset.BitSet
	second       bitset.BitSet
	lastMonthDay int  //在MonthDay模式代表每个月的倒数第几天,在WeekDay模式代表每个月的最后一个星期几
	useWeekDay   bool //切换MonthDay,WeekDay模式

	handler *timeHandler

	Head
}

// 一个timeNode节点有4个状态
// 1.存在于初始化链表中
// 2.被移动到tmp链表
// 3.1 和 3.2是if else的状态
// 	3.1被移动到new链表
// 	3.2直接执行
// 1和3.1状态是没有问题的
// 2和3.2状态会是没有锁保护下的操作,会有数据竞争
func (this *timeNode) Stop() {

	atomic.StoreUint32(&this.stop, haveStop)

	// 使用版本号算法让timeNode知道自己是否被移动了
	// timeNode的version和表头的version一样表示没有被移动可以直接删除
	// 如果不一样，可能在第2或者3.2状态，使用惰性删除
	cpyList := (*Time)(atomic.LoadPointer(&this.list))
	cpyList.Lock()
	defer cpyList.Unlock()
	if atomic.LoadUint64(&this.version) != atomic.LoadUint64(&cpyList.version) {
		return
	}

	cpyList.Del(&this.Head)

	this.handler.noders.Delete(this)
}

func (this *timeNode) GetDelay() time.Duration {
	return time.Duration(this.delay)
}

func (this *timeNode) GetInterval() time.Duration {
	return time.Duration(this.interval)
}

func (this *timeNode) GetCallback() func(TimeNoder) {
	return this.callback
}

//基础打点更新函数
func (this *timeNode) intervalExpireFunc() (uint64, bool) {
	if this.loopMax > 0 && this.loopCur >= this.loopMax {
		return this.interval, false
	}
	return this.interval, true
}

func parseWeekWords(s string) (time.Weekday, error) {
	switch s {
	case "SUN":
		return time.Sunday, nil
	case "MON":
		return time.Monday, nil
	case "TUE":
		return time.Tuesday, nil
	case "WED":
		return time.Wednesday, nil
	case "THD":
		return time.Thursday, nil
	case "FRI":
		return time.Friday, nil
	case "SAT":
		return time.Saturday, nil
	}
	return -1, log.Error("unrecognized week word:" + s)
}

//忽略检测,仅检测星期和月天
func ignoreCheck(s string) bool {
	return regexp.MustCompile(`\s*\?\s*`).MatchString(s)
}

//通配符检测
func everyCheck(s string) bool {
	return regexp.MustCompile(`\s*\*\s*`).FindString(s) == s
}

//纯数字检测
func digitalCheck(s string) int {
	if regexp.MustCompile(`\s*(\d+)\s*`).FindString(s) != s {
		return -1
	}
	ret, _ := strconv.Atoi(regexp.MustCompile(`\s*(\d+)\s*`).ReplaceAllString(s, "$1"))
	return ret
}

func rangeCheck(s string) (bool, []int) {
	r := regexp.MustCompile(`\s*([0-9]+)\s*\-\s*([0-9])+\s*`)
	if r.FindString(s) != s {
		return false, nil
	}
	ret := []int{}
	res := r.ReplaceAllString(s, "$1")
	split := strings.Split(res, "-")
	min, _ := strconv.Atoi(strings.TrimSpace(split[0]))
	max, _ := strconv.Atoi(strings.TrimSpace(split[0]))
	if min > max {
		log.Error("timewheel:parse error,range min must <= max " + s)
		return false, nil
	}
	for i := min; i < max; i++ {
		ret = append(ret, i)
	}
	return true, ret
}

//检测分隔符
func splitCheck(s string) (bool, []int) {
	r := regexp.MustCompile(`\s*([\d+\s*\,\s*]+\s*\d*)\s*`)
	res := r.FindString(s)
	if res == "" {
		return false, nil
	}
	ret := []int{}
	res = r.ReplaceAllString(s, "$1")
	splits := strings.Split(res, ",")
	for _, v := range splits {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		d, err := strconv.Atoi(v)
		if err != nil {
			log.Error(err.Error())
			return false, ret
		}
		ret = append(ret, d)
	}
	return true, ret
}

func checkMaxMin(period int, entry []int, offset int) bool {
	var max = xmath.MaxArr[int](entry) + offset
	var min = xmath.MinArr[int](entry) + offset
	if min < 0 {
		return false
	}
	if max > period-1 {
		return false
	}
	return true
}

func checkCronString(s string, period int) (bitset.BitSet, error) {
	if everyCheck(s) {
		switch period {
		case 60:
			return ALL_MINUTE, nil
		case 12:
			return ALL_MONTH, nil
		case 7:
			return ALL_WEEK, nil
		case 24:
			return ALL_HOUR, nil
		case 31:
			return ALL_DAY, nil
		default:
			return 0, log.Error("timewheel:period not exist")
		}
	}
	offset := 0
	if period == 7 || period == 31 || period == 12 {
		offset = -1
	}
	if d := digitalCheck(s); d != -1 {
		if !checkMaxMin(period, []int{d}, offset) {
			return 0, log.Error("timewheel:range error")
		}
		var ret bitset.BitSet = 0
		ret = ret.Set(d+offset, true)
		return ret, nil
	}
	if ok, entry := splitCheck(s); ok {
		if !checkMaxMin(period, entry, offset) {
			return 0, log.Error("timewheel:range error")
		}
		var ret bitset.BitSet = 0
		for _, v := range entry {
			ret = ret.Set(v+offset, true)
		}
		return ret, nil
	}
	if ok, entry := rangeCheck(s); ok {
		if !checkMaxMin(period, entry, offset) {
			return 0, log.Error("timewheel:range error")
		}
		var ret bitset.BitSet = 0
		for _, v := range entry {
			ret = ret.Set(v+offset, true)
		}
		return ret, nil
	}
	return 0, log.Error("timewheel:parse error")
}

func (this *timeNode) parseCron(second, minute, hour, day, month, weekday string) error {
	ignore_day := ignoreCheck(day)
	ignore_weekday := ignoreCheck(weekday)
	if ignore_day && ignore_weekday {
		return log.Error("timewheel:cant have 2 ?")
	}
	if !ignore_day && !ignore_weekday {
		return log.Error("timewheel:must have 1 ? in day or week day")
	}
	if ignore_day {
		b, err := checkCronString(weekday, 7)
		if err != nil {
			return err
		}
		this.useWeekDay = true
		this.weekday = b
	} else {
		b, err := checkCronString(day, 31)
		if err != nil {
			return err
		}
		this.useWeekDay = false
		this.monthday = b
	}
	b, err := checkCronString(month, 12)
	if err != nil {
		return err
	}
	b, err = checkCronString(second, 60)
	if err != nil {
		return err
	}
	this.second = b
	b, err = checkCronString(minute, 60)
	if err != nil {
		return err
	}
	this.minute = b
	b, err = checkCronString(hour, 24)
	if err != nil {
		return err
	}
	this.hour = b
	b, err = checkCronString(month, 12)
	if err != nil {
		return err
	}
	this.month = b
	return nil
}

func (this timeNode) cronExpireFunc(t *timeWheel) (uint64, bool) {
	now := t.Now()
	year := now.Year()
	monthday := now.Day()
	weekday := now.Weekday()
	month := int(now.Month())
	hour := now.Hour()
	minute := now.Minute()
	second := now.Second()
	move_year := false
	move_month := false
	move_day := false
	move_hour := false
	move_minute := false
	if second, move_minute = this.getNextSecond(second); move_minute {
		if minute, move_hour = this.getNextMinute(minute); move_hour {
			if hour, move_day = this.getNextHour(hour); move_day {
				if monthday, move_month = this.getNextDay(monthday, now, this.useWeekDay, int(weekday), this.lastMonthDay); move_month {
					month, move_year = this.getNextMonth(month)
				}
			}
		}
	}
	if move_year {
		year += 1
	}
	next_time := time.Date(year, time.Month(month), monthday, hour, minute, second, 0, time.Local)
	d := next_time.Sub(now)
	return uint64(d), true
}

func (this *timeNode) getNextMonth(month int) (next_month int, move_year bool) {
	month += 1
	for i := 0; i < 12; i++ {
		month += 1
		if month > 11 {
			month = 0
			move_year = true
		}
		if this.month.Get(month) {
			next_month = month
			return
		}
	}
	log.Panic("timewheel:month not found")
	return
}

func (this *timeNode) getNextWeekDay(day int, week_day int, now time.Time, last_day int) (next_day int, move_month bool) {
	days := util.GetDaysOfMonth(now)
	if last_day >= 0 {
		next_day = util.GetMonthDayByLastWeekDay(now, time.Weekday(week_day))
		if day+1 > next_day {
			return next_day, true
		}
		return next_day, false
	}
	for i := 0; i < 7; i++ {
		week_day += 1
		day += 1
		if day > days-1 {
			day = 0
			move_month = true
		}
		if this.weekday.Get(week_day) {
			next_day = day
			return
		}
	}
	log.Panic("timewheel:week_day not found")
	return
}

func (this *timeNode) getNextDay(day int, now time.Time, is_weekday bool, weekday int, last_day int) (next_day int, move_month bool) {
	if is_weekday {
		return this.getNextWeekDay(day, weekday, now, last_day)
	}
	days := util.GetDaysOfMonth(now)
	if last_day >= 0 {
		if day+1 > last_day {
			return last_day, true
		}
		return last_day, false
	}
	for i := 0; i < days; i++ {
		day += 1
		if day > days-1 {
			day = 0
			move_month = true
		}
		if this.monthday.Get(day) {
			next_day = day
			return
		}
	}
	log.Panic("timewheel:day not found")
	return
}

func (this *timeNode) getNextHour(hour int) (next_hour int, move_day bool) {
	for i := 0; i < 24; i++ {
		hour += 1
		if hour > 23 {
			hour = 0
			move_day = true
		}
		if this.hour.Get(hour) {
			next_hour = hour
			return
		}
	}
	log.Panic("timewheel:hour not found")
	return
}

func (this *timeNode) getNextMinute(minute int) (next_minute int, move_hour bool) {
	for i := 0; i < 60; i++ {
		minute += 1
		if minute > 59 {
			minute = 0
			move_hour = true
		}
		if this.minute.Get(minute) {
			next_minute = minute
			return
		}
	}
	log.Panic("timewheel:minute not found")
	return
}

func (this *timeNode) getNextSecond(second int) (next_second int, move_minute bool) {
	for i := 0; i < 60; i++ {
		second += 1
		if second > 59 {
			second = 0
			move_minute = true
		}
		if this.second.Get(second) {
			next_second = second
			return
		}
	}
	log.Panic("timewheel:second not found")
	return
}
