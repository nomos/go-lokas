package util

import (
	"github.com/noaway/dateparse"
	"time"
)

var Time2028 = time.Date(2028, 1, 1, 0, 0, 0, 0, time.Local)

func SameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
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

func UnixSecondToString(unix int64) string{
	return FormatTimeToString(time.Unix(unix,0))
}

func UnixNanoToString(nano int64) string{
	return FormatTimeToString(time.Unix(0,nano))
}

func MilliSecondToString(millisec int64) string{
	return FormatTimeToString(GetTimeFromMs(millisec))
}

func GetTimeFromMs(t int64)time.Time{
	return time.Unix(t/time.Millisecond.Nanoseconds(),0)
}

func CheckTimeFormat(src, layout string) bool {
	_, err := time.Parse(layout, src)
	return err == nil
}

func GetDateWithOffset(t time.Time,offset time.Duration) string {
	return GetTimeWithOffset(t,offset).Format("060102")
}

func GetTimeWithOffset(t time.Time,offset time.Duration) time.Time {
	return t.Add(-offset)
}

func GetDateNoOffset(t time.Time) string {
	return t.Format("060102")
}

//GetNowDateWithOffset
//得到带重置时间的当天的YYYYMMHH
func GetNowDateWithOffset(offset time.Duration) string {
	return GetDateWithOffset(time.Now(),offset)
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

func TimeParseLocal(str string)(time.Time,error){
	return dateparse.ParseLocal(str)
}

func FormatTimeToDayString(time time.Time)string {
	return time.Format("2006-01-02")
}

func FormatTimeToMonthString(time time.Time)string {
	return time.Format("2006-01")
}

func FormatTimeToYearString(time time.Time)string {
	return time.Format("2006")
}