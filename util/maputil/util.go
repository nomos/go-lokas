package maputil

func KeysOf[T1 comparable, T2 any](m map[T1]T2) []T1 {
	ret := []T1{}
	for k, _ := range m {
		ret = append(ret, k)
	}
	return ret
}
