package bitset

import "strconv"

type BitSet uint64

func (this BitSet) Values(len int) []int {
	if len > 64 {
		len = 64
	}
	ret := make([]int, 0, len)
	b := this
	for i := 0; i < len; i++ {
		if b&1 == 1 {
			ret = append(ret, i)
		}
		b = b >> 1
	}
	return ret
}

func (this BitSet) Get(index int) bool {
	return (this>>index)&1 == 1
}

func (this BitSet) Print() string {
	return strconv.FormatInt(int64(this), 2)
}

func (this BitSet) Set(index int, v bool) BitSet {
	if v {
		return this | (1 << index)
	} else {
		return this & (^(1 << index))
	}
}
