package ecs

import "github.com/nomos/go-lokas"

func GetComponentName(ecs lokas.IRuntime,c Component)string {
	return ""
}

func GetComponentSyncAble(ecs lokas.IRuntime,c interface{})bool {
	switch c.(type) {
	case Component:
		return ecs.IsSyncAble(GetComponentName(ecs,c.(Component)))
	case string:
		return ecs.IsSyncAble(c.(string))
	}
	return false
}

func HasEntityInSlice(slice []*Entity,e *Entity)bool {
	for _,v :=range slice{
		if v == e {
			return true
		}
	}
	return false
}

