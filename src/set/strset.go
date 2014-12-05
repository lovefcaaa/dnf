package set

import (
	"fmt"
	"sort"
	"strconv"
)

type StrSet struct {
	data []string // AdIndex数据量较小，使用slice比map更省gc开销
}

func NewStrSet() *StrSet {
	return &StrSet{
		data: make([]string, 0, 1),
	}
}

func (this *StrSet) Add(s string) {
	//pos := sort.StringSlice(this.data).Search(s)
	pos := sort.Search(len(this.data), func(i int) bool { return this.data[i] >= s })
	if pos >= len(this.data) || this.data[pos] != s {
		this.data = append(this.data, s)
		sort.StringSlice(this.data).Sort()
	}
}

func (this *StrSet) Len() int {
	return len(this.data)
}

func (this *StrSet) Data() []string {
	return this.data
}

func Test() {
	s := NewStrSet()
	for i := 10; i != 0; i-- {
		s.Add(strconv.Itoa(i))
	}
	for i := 0; i != 10; i++ {
		s.Add(strconv.Itoa(i))
	}
	for _, str := range s.data {
		fmt.Println(str)
	}
}
