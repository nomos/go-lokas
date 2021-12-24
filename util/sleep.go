package util

import "syscall"

func Sleep(ms int64) {
	var tv syscall.Timeval
	_ = syscall.Gettimeofday(&tv)

	startTick := int64(tv.Sec)*int64(1000000) + int64(tv.Usec) + ms*1000
	endTick := int64(0)
	for endTick < startTick {
		_ = syscall.Gettimeofday(&tv)
		endTick = int64(tv.Sec)*int64(1000000) + int64(tv.Usec)
	}
}

func SleepUtil(ms int64,f func()bool) {
	var tv syscall.Timeval
	_ = syscall.Gettimeofday(&tv)

	startTick := int64(tv.Sec)*int64(1000000) + int64(tv.Usec) + ms*1000
	endTick := int64(0)
	for endTick < startTick {
		if f() {
			break
		}
		_ = syscall.Gettimeofday(&tv)
		endTick = int64(tv.Sec)*int64(1000000) + int64(tv.Usec)
	}
}


func USleep(us int64) {
	var tv syscall.Timeval
	_ = syscall.Gettimeofday(&tv)

	startTick := int64(tv.Sec)*int64(1000000) + int64(tv.Usec) + us
	endTick := int64(0)
	for endTick < startTick {
		_ = syscall.Gettimeofday(&tv)
		endTick = int64(tv.Sec)*int64(1000000) + int64(tv.Usec)
	}
}
