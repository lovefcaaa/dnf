package dnf

import (
	"errors"
	"sort"

	"set"
)

type Cond struct {
	Key string
	Val string
}

func searchCondCheck(conds []Cond) error {
	if conds == nil || len(conds) == 0 {
		return errors.New("no conds to search")
	}
	m := make(map[string]bool)
	for _, cond := range conds {
		if _, ok := m[cond.Key]; ok {
			return errors.New("duplicate keys: " + cond.Key)
		}
		m[cond.Key] = true
	}
	return nil
}

func Search(conds []Cond) (docs []int, err error) {
	if err := searchCondCheck(conds); err != nil {
		return nil, err
	}
	termids := make([]int, 0)
	for i := 0; i < len(conds); i++ {
		if id, ok := termMap[conds[i].Key+"%"+conds[i].Val]; ok {
			termids = append(termids, id)
		}
	}
	if len(termids) == 0 {
		return nil, errors.New("All cond are not in inverse list")
	}
	return doSearch(termids), nil
}

func doSearch(terms []int) (docs []int) {
	conjs := getConjs(terms)
	if len(conjs) == 0 {
		return nil
	}
	return getDocs(conjs)
}

func getDocs(conjs []int) (docs []int) {
	conjRvsLock.RLock()
	defer conjRvsLock.RUnlock()

	set := set.NewIntSet()

	for _, conj := range conjs {
		ASSERT(conj < len(conjRvs))
		doclist := conjRvs[conj]
		if doclist == nil {
			continue
		}
		for _, doc := range doclist {
			inTime := false

			docs_.Lock()
			if docs_.docs[doc].attr.Tr.CoverToday() {
				inTime = true
			}
			docs_.Unlock()

			if inTime {
				set.Add(doc)
			}
		}
	}
	return set.ToSlice()
}

func getConjs(terms []int) (conjs []int) {
	conjSzRvsLock.RLock()
	defer conjSzRvsLock.RUnlock()

	n := len(terms)
	ASSERT(len(conjSzRvs) > 0)
	if n >= len(conjSzRvs) {
		n = len(conjSzRvs) - 1
	}

	conjSet := set.NewIntSet()

	for i := 0; i <= n; i++ {
		termlist := conjSzRvs[i]
		if termlist == nil || len(termlist) == 0 {
			continue
		}

		countSet := set.NewCountSet(i)

		for _, tid := range terms {
			idx := sort.Search(len(termlist), func(i int) bool {
				return termlist[i].termId >= tid
			})
			if idx < len(termlist) && termlist[idx].termId == tid &&
				termlist[idx].cList != nil {

				for _, pair := range termlist[idx].cList {
					countSet.Add(pair.conjId, pair.belong)
				}
			}
		}

		/* 处理∅ */
		if i == 0 {
			for _, pair := range termlist[0].cList {
				ASSERT(pair.belong == true)
				countSet.Add(pair.conjId, pair.belong)
			}
		}

		conjSet.AddSlice(countSet.ToSlice())
	}

	return conjSet.ToSlice()
}
