package hset

import (
	"encoding/json"
)

type Set interface {
	Add(items ...interface{})
	Remove(items ...interface{})
	Clear()
	Contains(items ...interface{}) bool
	Len() int
	Same(other Set) bool
	Values() []interface{}
	String() string
}

func (set *Hset) ToJSON() ([]byte, error) {
	return json.Marshal(set.Values())
}

func (set *Hset) FromJSON(data []byte) error {
	elements := []interface{}{}
	err := json.Unmarshal(data, &elements)
	if err == nil {
		set.Clear()
		set.Add(elements...)
	}

	return err
}
