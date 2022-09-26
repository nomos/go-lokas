package util

import (
	"github.com/noaway/dateparse"
	"time"
)

const dayDuration time.Duration = time.Hour * 24

var Time2028 = time.Date(2028, 1, 1, 0, 0, 0, 0, time.Local)

func SameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

//获取这星期的星期几
func GetWeekDay(date time.Time, weekday time.Weekday) time.Time {
	day_epoch := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	t := date.Weekday()
	return day_epoch.Add(dayDuration * time.Duration(weekday-t))
}
func GetMonthStart(t time.Time) time.Time {
	monthStartDay := t.AddDate(0, 0, -t.Day()+1)
	monthStartTime := time.Date(monthStartDay.Year(), monthStartDay.Month(), monthStartDay.Day(), 0, 0, 0, 0, t.Location())
	return monthStartTime
}

func GetDayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func GetYearStart(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

func GetNextYearStart(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location()).AddDate(1, 0, 0)
}

func GetNextDayStart(t time.Time) time.Time {
	dayStart := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return dayStart.AddDate(0, 0, 1)
}

func GetNextMonthsStart(t time.Time, months int) time.Time {
	monthStartDay := t.AddDate(0, 0, -t.Day()+1)
	monthStartTime := time.Date(monthStartDay.Year(), monthStartDay.Month(), monthStartDay.Day(), 0, 0, 0, 0, t.Location())
	return monthStartTime.AddDate(0, months, 0)
}

func GetNextMonthStart(t time.Time) time.Time {
	monthStartDay := t.AddDate(0, 0, -t.Day()+1)
	monthStartTime := time.Date(monthStartDay.Year(), monthStartDay.Month(), monthStartDay.Day(), 0, 0, 0, 0, t.Location())
	return monthStartTime.AddDate(0, 1, 0)
}

//获取本月的起始时间
func GetMonthStartEnd(t time.Time) (time.Time, time.Time) {
	monthStartDay := t.AddDate(0, 0, -t.Day()+1)
	monthStartTime := time.Date(monthStartDay.Year(), monthStartDay.Month(), monthStartDay.Day(), 0, 0, 0, 0, t.Location())
	monthEndDay := monthStartTime.AddDate(0, 1, -1)
	monthEndTime := time.Date(monthEndDay.Year(), monthEndDay.Month(), monthEndDay.Day(), 23, 59, 59, 0, t.Location())
	return monthStartTime, monthEndTime
}

//获取本月有几天
func GetDaysOfMonth(t time.Time) int {
	monthStartDay := t.AddDate(0, 0, -t.Day()+1)
	monthStartTime := time.Date(monthStartDay.Year(), monthStartDay.Month(), monthStartDay.Day(), 0, 0, 0, 0, t.Location())
	return monthStartTime.AddDate(0, 1, -1).Day() + 1
}

func GetLastDayOfMonth(t time.Time) time.Time {
	monthStartDay := t.AddDate(0, 0, -t.Day()+1)
	monthStartTime := time.Date(monthStartDay.Year(), monthStartDay.Month(), monthStartDay.Day(), 0, 0, 0, 0, t.Location())
	return monthStartTime.AddDate(0, 1, -1)
}

func GetLastDaysOfMonth(t time.Time, days int) time.Time {
	monthStartDay := t.AddDate(0, 0, -t.Day()+1)
	monthStartTime := time.Date(monthStartDay.Year(), monthStartDay.Month(), monthStartDay.Day(), 0, 0, 0, 0, t.Location())
	return monthStartTime.AddDate(0, 1, -days)
}

func GetLastWeekDay(t time.Time, weekday time.Weekday) time.Time {
	lastDay := GetLastDayOfMonth(t)
	lastDayWeekday := lastDay.Weekday()
	dayShift := weekday - lastDayWeekday
	if dayShift > 0 {
		dayShift -= 7
	}
	return lastDay.AddDate(0, 0, int(dayShift))
}

func GetMonthDayByLastWeekDay(t time.Time, weekday time.Weekday) int {
	lastDay := GetLastDayOfMonth(t)
	lastDayWeekday := lastDay.Weekday()
	dayShift := weekday - lastDayWeekday
	if dayShift > 0 {
		dayShift -= 7
	}
	return lastDay.Day() + int(dayShift)
}

func IsToday(ts int64) bool {
	return SameDay(time.Unix(ts, 0), time.Now())
}

func FormatTimeToString(time time.Time) string {
	return time.Format("2006-01-02 15:04:05")
}

func FormatTimeToISOString(time time.Time) string {
	return time.Format("2006-01-02T15:04:05")
}

func UnixSecondToString(unix int64) string {
	return FormatTimeToString(time.Unix(unix, 0))
}

func UnixNanoToString(nano int64) string {
	return FormatTimeToString(time.Unix(0, nano))
}

func MilliSecondToString(millisec int64) string {
	return FormatTimeToString(GetTimeFromMs(millisec))
}

func GetTimeFromMs(t int64) time.Time {
	return time.Unix(t/time.Millisecond.Nanoseconds(), 0)
}

func CheckTimeFormat(src, layout string) bool {
	_, err := time.Parse(layout, src)
	return err == nil
}

func GetDateWithOffset(t time.Time, offset time.Duration) string {
	return GetTimeWithOffset(t, offset).Format("060102")
}

func GetTimeWithOffset(t time.Time, offset time.Duration) time.Time {
	return t.Add(-offset)
}

func GetDateNoOffset(t time.Time) string {
	return t.Format("060102")
}

//GetNowDateWithOffset
//得到带重置时间的当天的YYYYMMHH
func GetNowDateWithOffset(offset time.Duration) string {
	return GetDateWithOffset(time.Now(), offset)
}

//得到当前时间到2028年的时间
//一般用在排行榜上,比如同等级,判断时间先后
func Time2TenYears() time.Duration {
	return Time2028.Sub(time.Now())
}

//等到tm1-tm2的天数，
func GetPassedDays(tm1, tm2 time.Time) int64 {
	return tm1.Unix()/86400 - tm2.Unix()/86400
}

func TimeParseLocal(str string) (time.Time, error) {
	return dateparse.ParseLocal(str)
}

func FormatTimeToDayString(time time.Time) string {
	return time.Format("2006-01-02")
}

func FormatTimeToMonthString(time time.Time) string {
	return time.Format("2006-01")
}

func FormatTimeToYearString(time time.Time) string {
	return time.Format("2006")
}
