package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
)


type Set interface {
	Add(i interface{}) bool

	Cardinality() int

	Clear()

	Clone() Set

	Contains(i ...interface{}) bool

	Difference(other Set) Set

	Equal(other Set) bool

	Intersect(other Set) Set

	IsProperSubset(other Set) bool

	IsProperSuperset(other Set) bool

	IsSubset(other Set) bool

	IsSuperset(other Set) bool

	Each(func(interface{}) bool)

	Iter() <-chan interface{}

	Iterator() *Iterator

	Remove(i interface{})

	String() string

	SymmetricDifference(other Set) Set

	Union(other Set) Set

	Pop() interface{}

	PowerSet() Set

	CartesianProduct(other Set) Set

	ToSlice() []interface{}
}

func NewSet(s ...interface{}) Set {
	set := newThreadSafeSet()
	for _, item := range s {
		set.Add(item)
	}
	return &set
}

func NewSetWith(elts ...interface{}) Set {
	return NewSetFromSlice(elts)
}

func NewSetFromSlice(s []interface{}) Set {
	a := NewSet(s...)
	return a
}

func NewThreadUnsafeSet() Set {
	set := newThreadUnsafeSet()
	return &set
}

func NewThreadUnsafeSetFromSlice(s []interface{}) Set {
	a := NewThreadUnsafeSet()
	for _, item := range s {
		a.Add(item)
	}
	return a
}

type threadSafeSet struct {
	s threadUnsafeSet
	sync.RWMutex
}

func newThreadSafeSet() threadSafeSet {
	return threadSafeSet{s: newThreadUnsafeSet()}
}

func (this *threadSafeSet) Add(i interface{}) bool {
	this.Lock()
	ret := this.s.Add(i)
	this.Unlock()
	return ret
}

func (this *threadSafeSet) Contains(i ...interface{}) bool {
	this.RLock()
	ret := this.s.Contains(i...)
	this.RUnlock()
	return ret
}

func (this *threadSafeSet) IsSubset(other Set) bool {
	o := other.(*threadSafeSet)

	this.RLock()
	o.RLock()

	ret := this.s.IsSubset(&o.s)
	this.RUnlock()
	o.RUnlock()
	return ret
}

func (this *threadSafeSet) IsProperSubset(other Set) bool {
	o := other.(*threadSafeSet)

	this.RLock()
	defer this.RUnlock()
	o.RLock()
	defer o.RUnlock()

	return this.s.IsProperSubset(&o.s)
}

func (this *threadSafeSet) IsSuperset(other Set) bool {
	return other.IsSubset(this)
}

func (this *threadSafeSet) IsProperSuperset(other Set) bool {
	return other.IsProperSubset(this)
}

func (this *threadSafeSet) Union(other Set) Set {
	o := other.(*threadSafeSet)

	this.RLock()
	o.RLock()

	unsafeUnion := this.s.Union(&o.s).(*threadUnsafeSet)
	ret := &threadSafeSet{s: *unsafeUnion}
	this.RUnlock()
	o.RUnlock()
	return ret
}

func (this *threadSafeSet) Intersect(other Set) Set {
	o := other.(*threadSafeSet)

	this.RLock()
	o.RLock()

	unsafeIntersection := this.s.Intersect(&o.s).(*threadUnsafeSet)
	ret := &threadSafeSet{s: *unsafeIntersection}
	this.RUnlock()
	o.RUnlock()
	return ret
}

func (this *threadSafeSet) Difference(other Set) Set {
	o := other.(*threadSafeSet)

	this.RLock()
	o.RLock()

	unsafeDifference := this.s.Difference(&o.s).(*threadUnsafeSet)
	ret := &threadSafeSet{s: *unsafeDifference}
	this.RUnlock()
	o.RUnlock()
	return ret
}

func (this *threadSafeSet) SymmetricDifference(other Set) Set {
	o := other.(*threadSafeSet)

	this.RLock()
	o.RLock()

	unsafeDifference := this.s.SymmetricDifference(&o.s).(*threadUnsafeSet)
	ret := &threadSafeSet{s: *unsafeDifference}
	this.RUnlock()
	o.RUnlock()
	return ret
}

func (this *threadSafeSet) Clear() {
	this.Lock()
	this.s = newThreadUnsafeSet()
	this.Unlock()
}

func (this *threadSafeSet) Remove(i interface{}) {
	this.Lock()
	delete(this.s, i)
	this.Unlock()
}

func (this *threadSafeSet) Cardinality() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.s)
}

func (this *threadSafeSet) Each(cb func(interface{}) bool) {
	this.RLock()
	for elem := range this.s {
		if cb(elem) {
			break
		}
	}
	this.RUnlock()
}

func (this *threadSafeSet) Iter() <-chan interface{} {
	ch := make(chan interface{})
	go func() {
		this.RLock()

		for elem := range this.s {
			ch <- elem
		}
		close(ch)
		this.RUnlock()
	}()

	return ch
}

func (this *threadSafeSet) Iterator() *Iterator {
	iterator, ch, stopCh := newIterator()

	go func() {
		this.RLock()
	L:
		for elem := range this.s {
			select {
			case <-stopCh:
				break L
			case ch <- elem:
			}
		}
		close(ch)
		this.RUnlock()
	}()

	return iterator
}

func (this *threadSafeSet) Equal(other Set) bool {
	o := other.(*threadSafeSet)

	this.RLock()
	o.RLock()

	ret := this.s.Equal(&o.s)
	this.RUnlock()
	o.RUnlock()
	return ret
}

func (this *threadSafeSet) Clone() Set {
	this.RLock()

	unsafeClone := this.s.Clone().(*threadUnsafeSet)
	ret := &threadSafeSet{s: *unsafeClone}
	this.RUnlock()
	return ret
}

func (this *threadSafeSet) String() string {
	this.RLock()
	ret := this.s.String()
	this.RUnlock()
	return ret
}

func (this *threadSafeSet) PowerSet() Set {
	this.RLock()
	unsafePowerSet := this.s.PowerSet().(*threadUnsafeSet)
	this.RUnlock()

	ret := &threadSafeSet{s: newThreadUnsafeSet()}
	for subset := range unsafePowerSet.Iter() {
		unsafeSubset := subset.(*threadUnsafeSet)
		ret.Add(&threadSafeSet{s: *unsafeSubset})
	}
	return ret
}

func (this *threadSafeSet) Pop() interface{} {
	this.Lock()
	defer this.Unlock()
	return this.s.Pop()
}

func (this *threadSafeSet) CartesianProduct(other Set) Set {
	o := other.(*threadSafeSet)

	this.RLock()
	o.RLock()

	unsafeCartProduct := this.s.CartesianProduct(&o.s).(*threadUnsafeSet)
	ret := &threadSafeSet{s: *unsafeCartProduct}
	this.RUnlock()
	o.RUnlock()
	return ret
}

func (this *threadSafeSet) ToSlice() []interface{} {
	keys := make([]interface{}, 0, this.Cardinality())
	this.RLock()
	for elem := range this.s {
		keys = append(keys, elem)
	}
	this.RUnlock()
	return keys
}

func (this *threadSafeSet) MarshalJSON() ([]byte, error) {
	this.RLock()
	b, err := this.s.MarshalJSON()
	this.RUnlock()

	return b, err
}

func (this *threadSafeSet) UnmarshalJSON(p []byte) error {
	this.RLock()
	err := this.s.UnmarshalJSON(p)
	this.RUnlock()

	return err
}

type threadUnsafeSet map[interface{}]struct{}

// An OrderedPair represents a 2-tuple of values.
type OrderedPair struct {
	First  interface{}
	Second interface{}
}

func newThreadUnsafeSet() threadUnsafeSet {
	return make(threadUnsafeSet)
}

// Equal says whether two 2-tuples contain the same values in the same order.
func (pair *OrderedPair) Equal(other OrderedPair) bool {
	if pair.First == other.First &&
		pair.Second == other.Second {
		return true
	}

	return false
}

func (this *threadUnsafeSet) Add(i interface{}) bool {
	_, found := (*this)[i]
	if found {
		return false //False if it existed already
	}

	(*this)[i] = struct{}{}
	return true
}

func (this *threadUnsafeSet) Contains(i ...interface{}) bool {
	for _, val := range i {
		if _, ok := (*this)[val]; !ok {
			return false
		}
	}
	return true
}

func (this *threadUnsafeSet) IsSubset(other Set) bool {
	_ = other.(*threadUnsafeSet)
	if this.Cardinality() > other.Cardinality() {
		return false
	}
	for elem := range *this {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

func (this *threadUnsafeSet) IsProperSubset(other Set) bool {
	return this.IsSubset(other) && !this.Equal(other)
}

func (this *threadUnsafeSet) IsSuperset(other Set) bool {
	return other.IsSubset(this)
}

func (this *threadUnsafeSet) IsProperSuperset(other Set) bool {
	return this.IsSuperset(other) && !this.Equal(other)
}

func (this *threadUnsafeSet) Union(other Set) Set {
	o := other.(*threadUnsafeSet)

	unionedSet := newThreadUnsafeSet()

	for elem := range *this {
		unionedSet.Add(elem)
	}
	for elem := range *o {
		unionedSet.Add(elem)
	}
	return &unionedSet
}

func (this *threadUnsafeSet) Intersect(other Set) Set {
	o := other.(*threadUnsafeSet)

	intersection := newThreadUnsafeSet()
	// loop over smaller set
	if this.Cardinality() < other.Cardinality() {
		for elem := range *this {
			if other.Contains(elem) {
				intersection.Add(elem)
			}
		}
	} else {
		for elem := range *o {
			if this.Contains(elem) {
				intersection.Add(elem)
			}
		}
	}
	return &intersection
}

func (this *threadUnsafeSet) Difference(other Set) Set {
	_ = other.(*threadUnsafeSet)

	difference := newThreadUnsafeSet()
	for elem := range *this {
		if !other.Contains(elem) {
			difference.Add(elem)
		}
	}
	return &difference
}

func (this *threadUnsafeSet) SymmetricDifference(other Set) Set {
	_ = other.(*threadUnsafeSet)

	aDiff := this.Difference(other)
	bDiff := other.Difference(this)
	return aDiff.Union(bDiff)
}

func (this *threadUnsafeSet) Clear() {
	*this = newThreadUnsafeSet()
}

func (this *threadUnsafeSet) Remove(i interface{}) {
	delete(*this, i)
}

func (this *threadUnsafeSet) Cardinality() int {
	return len(*this)
}

func (this *threadUnsafeSet) Each(cb func(interface{}) bool) {
	for elem := range *this {
		if cb(elem) {
			break
		}
	}
}

func (this *threadUnsafeSet) Iter() <-chan interface{} {
	ch := make(chan interface{})
	go func() {
		for elem := range *this {
			ch <- elem
		}
		close(ch)
	}()

	return ch
}

func (this *threadUnsafeSet) Iterator() *Iterator {
	iterator, ch, stopCh := newIterator()

	go func() {
	L:
		for elem := range *this {
			select {
			case <-stopCh:
				break L
			case ch <- elem:
			}
		}
		close(ch)
	}()

	return iterator
}

func (this *threadUnsafeSet) Equal(other Set) bool {
	_ = other.(*threadUnsafeSet)

	if this.Cardinality() != other.Cardinality() {
		return false
	}
	for elem := range *this {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

func (this *threadUnsafeSet) Clone() Set {
	clonedSet := newThreadUnsafeSet()
	for elem := range *this {
		clonedSet.Add(elem)
	}
	return &clonedSet
}

func (this *threadUnsafeSet) String() string {
	items := make([]string, 0, len(*this))

	for elem := range *this {
		items = append(items, fmt.Sprintf("%v", elem))
	}
	return fmt.Sprintf("Set{%s}", strings.Join(items, ", "))
}

// String outputs a 2-tuple in the form "(A, B)".
func (pair OrderedPair) String() string {
	return fmt.Sprintf("(%v, %v)", pair.First, pair.Second)
}

func (this *threadUnsafeSet) Pop() interface{} {
	for item := range *this {
		delete(*this, item)
		return item
	}
	return nil
}

func (this *threadUnsafeSet) PowerSet() Set {
	powSet := NewThreadUnsafeSet()
	nullset := newThreadUnsafeSet()
	powSet.Add(&nullset)

	for es := range *this {
		u := newThreadUnsafeSet()
		j := powSet.Iter()
		for er := range j {
			p := newThreadUnsafeSet()
			if reflect.TypeOf(er).Name() == "" {
				k := er.(*threadUnsafeSet)
				for ek := range *(k) {
					p.Add(ek)
				}
			} else {
				p.Add(er)
			}
			p.Add(es)
			u.Add(&p)
		}

		powSet = powSet.Union(&u)
	}

	return powSet
}

func (this *threadUnsafeSet) CartesianProduct(other Set) Set {
	o := other.(*threadUnsafeSet)
	cartProduct := NewThreadUnsafeSet()

	for i := range *this {
		for j := range *o {
			elem := OrderedPair{First: i, Second: j}
			cartProduct.Add(elem)
		}
	}

	return cartProduct
}

func (this *threadUnsafeSet) ToSlice() []interface{} {
	keys := make([]interface{}, 0, this.Cardinality())
	for elem := range *this {
		keys = append(keys, elem)
	}

	return keys
}

// MarshalJSON creates a JSON array from the set, it marshals all elements
func (this *threadUnsafeSet) MarshalJSON() ([]byte, error) {
	items := make([]string, 0, this.Cardinality())

	for elem := range *this {
		b, err := json.Marshal(elem)
		if err != nil {
			return nil, err
		}

		items = append(items, string(b))
	}

	return []byte(fmt.Sprintf("[%s]", strings.Join(items, ","))), nil
}

// UnmarshalJSON recreates a set from a JSON array, it only decodes
// primitive types. Numbers are decoded as json.Number.
func (this *threadUnsafeSet) UnmarshalJSON(b []byte) error {
	var i []interface{}

	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	err := d.Decode(&i)
	if err != nil {
		return err
	}

	for _, v := range i {
		switch t := v.(type) {
		case []interface{}, map[string]interface{}:
			continue
		default:
			this.Add(t)
		}
	}

	return nil
}
