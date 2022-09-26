package maputil

func KeysOf[T1, T2](m map[T1]T2) []T1 {
	ret := []T1{}
	for k, _ := range m {
		ret = append(ret, k)
	}
	return ret
}
