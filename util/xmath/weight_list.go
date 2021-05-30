package xmath

import "math/rand"

func WeightList(list []int)int {
	weight:=0
	for _,v:=range list {
		weight+=v
	}
	r:=rand.Intn(weight)
	weight= 0
	for index,v:=range list {
		weight+=v
		if weight>=r {
			return index
		}
	}
	return 0
}