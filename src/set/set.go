package set

import (
	"sort"
)

type CountSet struct {
	count    int
	positive map[int]int
	negetive map[int]int
	result   map[int]bool
}

func NewCountSet(count int) *CountSet {
	return &CountSet{
		count:    count,
		positive: make(map[int]int),
		negetive: make(map[int]int),
		result:   make(map[int]bool),
	}
}

func (set *CountSet) Add(id int, post bool) {
	if !post {
		set.negetive[id] = 1
	} else {
		val := set.positive[id]
		val++
		if val >= set.count {
			set.result[id] = true
		} else {
			set.positive[id] = val
		}
	}
}

func (set *CountSet) ToSlice() []int {
	for k, _ := range set.negetive {
		if _, ok := set.result[k]; ok {
			delete(set.result, k)
		}
	}
	rc := make([]int, 0, len(set.result))
	for k, _ := range set.result {
		rc = append(rc, k)
	}
	if !sort.IntsAreSorted(rc) {
		sort.IntSlice(rc).Sort()
	}
	return rc
}

type IntSet struct {
	data map[int]bool
}

func NewIntSet() *IntSet {
	return &IntSet{data: make(map[int]bool)}
}

func (set *IntSet) Add(elem int) {
	set.data[elem] = true
}

func (set *IntSet) AddSlice(elems []int) {
	for _, elem := range elems {
		set.Add(elem)
	}
}

func (set *IntSet) ToSlice() []int {
	rc := make([]int, 0, len(set.data))
	for k, _ := range set.data {
		rc = append(rc, k)
	}
	if !sort.IntsAreSorted(rc) {
		sort.IntSlice(rc).Sort()
	}
	return rc
}
