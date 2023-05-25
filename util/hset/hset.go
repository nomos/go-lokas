package hset

import (
	"fmt"
	"strings"
	"sync"
)

type Hset struct {
	items map[interface{}]struct{}
	sync.RWMutex
}

var itemExists = struct{}{}

func New() *Hset {
	return &Hset{items: make(map[interface{}]struct{})}
}

func (hset *Hset) Add(items ...interface{}) {
	hset.Lock()
	for _, item := range items {
		hset.items[item] = itemExists
	}
	hset.Unlock()
}

func (hset *Hset) Remove(items ...interface{}) {
	hset.Lock()
	for _, item := range items {
		delete(hset.items, item)
	}
	hset.Unlock()
}

func (hset *Hset) Contains(items ...interface{}) bool {
	for _, item := range items {
		if _, contains := hset.items[item]; !contains {
			return false
		}
	}
	return true
}

func (hset *Hset) Clear() {
	hset.Lock()
	hset.items = make(map[interface{}]struct{})
	hset.Unlock()
}

func (hset *Hset) Exists(item interface{}) bool {
	hset.RLock()
	_, ok := hset.items[item]
	hset.RUnlock()

	return ok
}

func (hset *Hset) Empty() bool {
	return hset.Len() == 0
}

func (hset *Hset) Len() int {
	hset.RLock()
	size := len(hset.items)
	hset.RUnlock()
	return size
}

func (hset *Hset) Values() []interface{} {
	hset.RLock()
	defer hset.RUnlock()

	values := make([]interface{}, hset.Len())
	count := 0
	for item := range hset.items {
		values[count] = item
		count++
	}
	return values
}

func (hset *Hset) Same(other Set) bool {
	hset.RLock()
	defer hset.RUnlock()

	if other == nil {
		return false
	}
	if hset.Len() != other.Len() {
		return false
	}

	for key := range hset.items {
		if !other.Contains(key) {
			return false
		}
	}
	return true
}

func (hset *Hset) String(pre ...bool) string {
	str := ""
	if len(pre) > 0 {
		str = "Has Hset:\n"
	}
	items := []string{}
	for k := range hset.items {
		items = append(items, fmt.Sprintf("%v", k))
	}

	str += strings.Join(items, ", ")
	return str
}
