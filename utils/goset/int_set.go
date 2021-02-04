package goset

import (
	"sort"
	"sync"
)

type IntSet struct {
	sync.RWMutex
	m map[int]bool
}

// 新建集合对象
func NewIntSet(items ...int) *IntSet {
	s := &IntSet{
		m: make(map[int]bool, len(items)),
	}
	s.Add(items...)
	return s
}

// 添加元素
func (s *IntSet) Add(items ...int) {
	s.Lock()
	defer s.Unlock()
	for _, v := range items {
		s.m[v] = true
	}
}

// 删除元素
func (s *IntSet) Remove(items ...int) {
	s.Lock()
	defer s.Unlock()
	for _, v := range items {
		delete(s.m, v)
	}
}

// 判断元素是否存在
func (s *IntSet) Has(items ...int) bool {
	s.RLock()
	defer s.RUnlock()
	for _, v := range items {
		if _, ok := s.m[v]; !ok {
			return false
		}
	}
	return true
}

// 元素个数
func (s *IntSet) Count() int {
	return len(s.m)
}

// 清空集合
func (s *IntSet) Clear() {
	s.Lock()
	defer s.Unlock()
	s.m = map[int]bool{}
}

// 空集合判断
func (s *IntSet) Empty() bool {
	return len(s.m) == 0
}

// 无序列表
func (s *IntSet) List() []int {
	s.RLock()
	defer s.RUnlock()
	list := make([]int, 0, len(s.m))
	for item := range s.m {
		list = append(list, item)
	}
	return list
}

// 排序列表
func (s *IntSet) SortList() []int {
	s.RLock()
	defer s.RUnlock()
	list := make([]int, 0, len(s.m))
	for item := range s.m {
		list = append(list, item)
	}
	sort.Ints(list)
	return list
}

// 并集
func (s *IntSet) Union(sets ...*IntSet) *IntSet {
	r := NewIntSet(s.List()...)
	for _, set := range sets {
		for e := range set.m {
			r.m[e] = true
		}
	}
	return r
}

// 差集
func (s *IntSet) Minus(sets ...*IntSet) *IntSet {
	r := NewIntSet(s.List()...)
	for _, set := range sets {
		for e := range set.m {
			if _, ok := s.m[e]; ok {
				delete(r.m, e)
			}
		}
	}
	return r
}

// 交集
func (s *IntSet) Intersect(sets ...*IntSet) *IntSet {
	r := NewIntSet(s.List()...)
	for _, set := range sets {
		for e := range s.m {
			if _, ok := set.m[e]; !ok {
				delete(r.m, e)
			}
		}
	}
	return r
}

// 补集
func (s *IntSet) Complement(full *IntSet) *IntSet {
	r := NewIntSet()
	for e := range full.m {
		if _, ok := s.m[e]; !ok {
			r.Add(e)
		}
	}
	return r
}

//slice并集
func IntUnion(slice1, slice2 []int) *IntSet {
	return NewIntSet(slice1...).Union(NewIntSet(slice2...))
}

//slice差集
func IntMinus(slice1, slice2 []int) *IntSet {
	return NewIntSet(slice1...).Minus(NewIntSet(slice2...))
}

//slice交集
func IntIntersect(slice1, slice2 []int) *IntSet {
	return NewIntSet(slice1...).Intersect(NewIntSet(slice2...))
}

//slice补集
func IntComplement(slice1, slice2 []int) *IntSet {
	return NewIntSet(slice1...).Complement(NewIntSet(slice2...))
}
