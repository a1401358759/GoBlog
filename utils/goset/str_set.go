package goset

import (
	"sort"
	"sync"
)

type StrSet struct {
	sync.RWMutex
	m map[string]bool
}

// 新建集合对象
func NewStrSet(items ...string) *StrSet {
	s := &StrSet{
		m: make(map[string]bool, len(items)),
	}
	s.Add(items...)
	return s
}

// 添加元素
func (s *StrSet) Add(items ...string) {
	s.Lock()
	defer s.Unlock()
	for _, v := range items {
		s.m[v] = true
	}
}

// 删除元素
func (s *StrSet) Remove(items ...string) {
	s.Lock()
	defer s.Unlock()
	for _, v := range items {
		delete(s.m, v)
	}
}

// 判断元素是否存在
func (s *StrSet) Has(items ...string) bool {
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
func (s *StrSet) Count() int {
	return len(s.m)
}

// 清空集合
func (s *StrSet) Clear() {
	s.Lock()
	defer s.Unlock()
	s.m = map[string]bool{}
}

// 空集合判断
func (s *StrSet) Empty() bool {
	return len(s.m) == 0
}

// 无序列表
func (s *StrSet) List() []string {
	s.RLock()
	defer s.RUnlock()
	list := make([]string, 0, len(s.m))
	for item := range s.m {
		list = append(list, item)
	}
	return list
}

// 排序列表
func (s *StrSet) SortList() []string {
	s.RLock()
	defer s.RUnlock()
	list := make([]string, 0, len(s.m))
	for item := range s.m {
		list = append(list, item)
	}
	sort.Strings(list)
	return list
}

// 并集
func (s *StrSet) Union(sets ...*StrSet) *StrSet {
	r := NewStrSet(s.List()...)
	for _, set := range sets {
		for e := range set.m {
			r.m[e] = true
		}
	}
	return r
}

// 差集
func (s *StrSet) Minus(sets ...*StrSet) *StrSet {
	r := NewStrSet(s.List()...)
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
func (s *StrSet) Intersect(sets ...*StrSet) *StrSet {
	r := NewStrSet(s.List()...)
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
func (s *StrSet) Complement(full *StrSet) *StrSet {
	r := NewStrSet()
	for e := range full.m {
		if _, ok := s.m[e]; !ok {
			r.Add(e)
		}
	}
	return r
}

//slice并集
func StrUnion(slice1, slice2 []string) *StrSet {
	return NewStrSet(slice1...).Union(NewStrSet(slice2...))
}

//slice差集
func StrMinus(slice1, slice2 []string) *StrSet {
	return NewStrSet(slice1...).Minus(NewStrSet(slice2...))
}

//slice交集
func StrIntersect(slice1, slice2 []string) *StrSet {
	return NewStrSet(slice1...).Intersect(NewStrSet(slice2...))
}

//slice补集
func StrComplement(slice1, slice2 []string) *StrSet {
	return NewStrSet(slice1...).Complement(NewStrSet(slice2...))
}
