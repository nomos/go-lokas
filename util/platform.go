package util

import "runtime"

func IsOSX()bool {
	return IsDarwin()
}

func IsDarwin()bool {
	return runtime.GOOS=="darwin"
}

func IsWindows()bool {
	return runtime.GOOS=="windows"
}

func IsLinux()bool{
	return runtime.GOOS=="linux"
}

func IsAndroid()bool{
	return runtime.GOOS=="android"
}

func IsIOS()bool{
	return runtime.GOOS=="ios"
}

func IsWeb()bool{
	return runtime.GOOS=="js"
}

func IsFreeBSD()bool{
	return runtime.GOOS=="freebsd"
}
