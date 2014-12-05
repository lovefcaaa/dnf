package dnf

import (
	"errors"
	"sort"
	"sync"
)

var once sync.Once

func Init() {
	once.Do(func() {
		debugInit()

		/* post list */
		docs_ = &docList{docs: make([]Doc, 0, 16)}
		conjs_ = &conjList{conjs: make([]Conj, 0, 16)}
		amts_ = &amtList{amts: make([]Amt, 0, 16)}
		terms_ = &termList{terms: make([]Term, 0, 16)}
		termMap = make(map[string]int)

		/* empty set ∅ */
		terms_.terms = append(terms_.terms, Term{id: 0, key: "", val: ""})

		/* reverse list */
		conjRvs = make([][]int, 0)
		conjSzRvs = make([][]termRvs, 16)

		/* empty set ∅ reverse list */
		termrvslist := make([]termRvs, 0, 1)
		termrvslist = append(termrvslist, termRvs{termId: 0, cList: make([]cPair, 0)})
		conjSzRvs[0] = termrvslist
	})
}

/* add new doc and insert infos into reverse lists */
func AddDoc(name string, docid string, dnfDesc string) error {
	for _, doc := range docs_.docs {
		if doc.docid == docid {
			return errors.New("doc " + docid + "has been added before")
		}
	}
	if err := DnfCheck(dnfDesc); err != nil {
		return err
	}
	doAddDoc(name, docid, dnfDesc)
	return nil
}

func doAddDoc(name string, docid string, dnf string) {
	doc := &Doc{
		docid: docid,
		name:  name,
		dnf:   dnf,
		conjs: make([]int, 0),
	}

	var conjId int
	var orStr string

	i := skipSpace(&dnf, 0)
	for {
		i, conjId = conjParse(&dnf, i)
		doc.conjs = append(doc.conjs, conjId)
		i = skipSpace(&dnf, i+1)
		if i >= len(dnf) {
			break
		}
		orStr, i = getString(&dnf, i)
		ASSERT(orStr == "or")
		i = skipSpace(&dnf, i+1)
	}
	docInternalId := docs_.Add(doc)
	conjReverse1(docInternalId, doc.conjs)
}

/*
conj: ( age in {3, 4} and state not in {CA, NY })
*/
func conjParse(dnf *string, i int) (endIndex int, conjId int) {
	var key, val string
	var vals []string
	var belong bool
	var op string /* "in" or "not in" */

	conj := &Conj{amts: make([]int, 0)}

	ASSERT((*dnf)[i] == '(')

	for {
		/* get assignment key */
		i = skipSpace(dnf, i+1)
		key, i = getString(dnf, i)

		/* get assignment op */
		i = skipSpace(dnf, i)
		op, i = getString(dnf, i)
		if op == "in" {
			belong = true
		} else {
			ASSERT(op == "not")
			i = skipSpace(dnf, i)
			op, i = getString(dnf, i)
			ASSERT(op == "in")
			belong = false
		}

		/* get assignment vals */
		i = skipSpace(dnf, i)
		ASSERT((*dnf)[i] == '{')
		vals = make([]string, 0, 1)
		for {
			i = skipSpace(dnf, i+1)
			val, i = getString(dnf, i)
			vals = append(vals, val)
			i = skipSpace(dnf, i)
			if (*dnf)[i] == '}' {
				break
			}
			ASSERT((*dnf)[i] == ',')
		}
		amtId := amtBuild(key, vals, belong)
		conj.amts = append(conj.amts, amtId)
		if belong {
			conj.size++
		}

		/* get next assignment or end of this conjunction */
		i = skipSpace(dnf, i+1)
		if (*dnf)[i] == ')' {
			conjId = conjs_.Add(conj)
			endIndex = i

			/* reverse list insert */
			conjReverse2(conj)
			return
		}

		val, i = getString(dnf, i)
		ASSERT(val == "and")
	}
}

func amtBuild(key string, vals []string, belong bool) (amtId int) {
	amt := &Amt{terms: make([]int, 0), belong: belong}
	for _, val := range vals {
		term := &Term{key: key, val: val}
		tid := terms_.Add(term)
		amt.terms = append(amt.terms, tid)
	}
	return amts_.Add(amt)
}

/*
   doc: (age ∈ { 3, 4 } and state ∈ { NY } ) or ( state ∈ { CA } and gender ∈ { M } ) -->

       conj1: (age ∈ { 3, 4 } and state ∈ { NY } )
       conj2: ( state ∈ { CA } and gender ∈ { M } )
*/
type Doc struct {
	id         int    /* unique id */
	docid      string /* sent by doc adder */
	name       string /* name of doc, for ad management */
	dnf        string /* dnf decription */
	conjSorted bool   /* is conjs slice sorted? */
	conjs      []int  /* conjunction ids */
	active     bool
}

/*
   conjunction: age ∈ { 3, 4 } and state ∈ { NY } -->

       assignment1: age ∈ { 3, 4 }
       assignment2: state ∈ { NY }
*/
type Conj struct {
	id        int   /* unique id */
	size      int   /* conj size: number of ∈ */
	amtSorted bool  /* is amts slice sorted? */
	amts      []int /* assignments ids */
}

func (this *Conj) Equal(c *Conj) bool {
	if !this.amtSorted {
		sort.IntSlice(this.amts).Sort()
		this.amtSorted = true
	}
	if !c.amtSorted {
		sort.IntSlice(c.amts).Sort()
		c.amtSorted = true
	}
	if this.size != c.size {
		return false
	}
	if len(this.amts) != len(c.amts) {
		return false
	}
	for i, amtId := range this.amts {
		if amtId != c.amts[i] {
			return false
		}
	}
	return true
}

/*
   assignment: age ∈ { 3, 4 } -->

       term1: age ∈ { 3 }
       term2: age ∈ { 4 }
*/
type Amt struct {
	id         int   /* unique id */
	belong     bool  /* ∈ or ∉ */
	termSorted bool  /* is terms slice sorted? */
	terms      []int /* terms ids */
}

func (this *Amt) Equal(amt *Amt) bool {
	if !this.termSorted {
		sort.IntSlice(this.terms).Sort()
		this.termSorted = false
	}
	if !amt.termSorted {
		sort.IntSlice(amt.terms).Sort()
		amt.termSorted = false
	}
	if len(this.terms) != len(amt.terms) {
		return false
	}
	if this.belong != amt.belong {
		return false
	}
	for i, term := range this.terms {
		if term != amt.terms[i] {
			return false
		}
	}
	return true
}

/*
   term: state ∉ { CA }
   Term{id: xxx, key: state, val: CA, belong: false}
*/
type Term struct {
	id  int
	key string
	val string
}

func (this *Term) Equal(term *Term) bool {
	if this.key == term.key &&
		this.val == term.val {
		return true
	}
	return false
}

/* post lists */
type docList struct {
	mutex sync.RWMutex
	docs  []Doc
}

type conjList struct {
	mutex sync.RWMutex
	conjs []Conj
}

type amtList struct {
	mutex sync.RWMutex
	amts  []Amt
}

type termList struct {
	mutex sync.RWMutex
	terms []Term
}

var termMap map[string]int

var docs_ *docList
var conjs_ *conjList
var amts_ *amtList
var terms_ *termList

func (this *docList) Add(doc *Doc) int {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	doc.id = len(this.docs)
	doc.active = true
	if !doc.conjSorted {
		sort.IntSlice(doc.conjs).Sort()
		doc.conjSorted = true
	}
	this.docs = append(this.docs, *doc)
	return doc.id
}

func (this *conjList) Add(conj *Conj) (conjId int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	for i, c := range this.conjs {
		if c.Equal(conj) {
			conj.id = c.id
			return i
		}
	}
	conj.id = len(this.conjs)

	/* append post list */
	this.conjs = append(this.conjs, *conj)

	/* append reverse list */
	conjRvs = append(conjRvs, make([]int, 0))

	return conj.id
}

func (this *amtList) Add(amt *Amt) (amtId int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	for i, a := range this.amts {
		if a.Equal(amt) {
			amt.id = a.id
			return i
		}
	}
	amt.id = len(this.amts)
	this.amts = append(this.amts, *amt)
	return amt.id
}

func (this *termList) Add(term *Term) (termId int) {
	if id, ok := termMap[term.key+"%"+term.val]; ok {
		term.id = id
		return id
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()
	// for i, t := range this.terms {
	// 	if t.Equal(term) {
	// 		term.id = t.id
	// 		return i
	// 	}
	// }
	term.id = len(this.terms)
	this.terms = append(this.terms, *term)
	termMap[term.key+"%"+term.val] = term.id
	return term.id
}

/* reverse lists 1 */
/*
             | <-- sizeof conjs_ --> |
   conjRvs:  +--+--+--+--+--+--+--+--+
             |0 |1 |2 | ...    ...   |
             +--+--+--+--+--+--+--+--+
                 |
                 +--> doc1.id --> doc3.id --> docN.id
*/
var conjRvsLock sync.RWMutex
var conjRvs [][]int

/* build the first layer reverse list */
func conjReverse1(docId int, conjIds []int) {
	conjRvsLock.Lock()
	defer conjRvsLock.Unlock()
	rvsLen := len(conjRvs)
	for _, conjId := range conjIds {
		ASSERT(rvsLen > conjId)
		rvsDocList := conjRvs[conjId]

		/* append docId to rvsDocList and promise rvsDocList sorted */
		pos := sort.IntSlice(rvsDocList).Search(docId)
		if pos < len(rvsDocList) && rvsDocList[pos] == docId {
			/* doc id exists */
			return
		}

		rvsDocList = append(rvsDocList, docId)
		if len(rvsDocList) > 1 {
			if docId < rvsDocList[len(rvsDocList)-2] {
				sort.IntSlice(rvsDocList).Sort()
			}
		}
		conjRvs[conjId] = rvsDocList
	}
}

/* reverse lists 2 */
/*
                 +----- sizeof (conj)
                 |
 conjSzRvs:  +--+--+--+--+--+--+
             |0 |1 | ...  ...  |
             +--+--+--+--+--+--+
                 |
                 +--> +-------+-------+-------+-------+
                      |termId |termId |termId |termId |
      []termRvs:      +-------+-------+-------+-------+
                      | clist | clist | clist | clist |
                      +-------+-------+-------+-------+
                         |
                         +--> +-----+-----+-----+-----+-----+
                              |cId:1|cId:4|cId:4|cId:8|cId:9|
              []cPair:        +-----+-----+-----+-----+-----+
                              |  ∈  |  ∈  |  ∉  |  ∉  |  ∈  |
                              +-----+-----+-----+-----+-----+
*/
type cPair struct {
	conjId int
	belong bool
}

/* for sort interface */
type cPairSlice []cPair

func (p cPairSlice) Len() int { return len(p) }
func (p cPairSlice) Less(i, j int) bool {
	if p[i].conjId == p[j].conjId {
		return p[i].belong
	}
	return p[i].conjId < p[j].conjId
}
func (p cPairSlice) Swap(i, j int) {
	p[i].conjId, p[j].conjId = p[j].conjId, p[i].conjId
	p[i].belong, p[j].belong = p[j].belong, p[i].belong
}

type termRvs struct {
	termId int
	cList  []cPair
}

/* for sort interface */
type termRvsSlice []termRvs

func (p termRvsSlice) Len() int           { return len(p) }
func (p termRvsSlice) Less(i, j int) bool { return p[i].termId < p[j].termId }
func (p termRvsSlice) Swap(i, j int) {
	p[i].termId, p[j].termId = p[j].termId, p[i].termId
	p[i].cList, p[j].cList = p[j].cList, p[i].cList
}

var conjSzRvsLock sync.RWMutex
var conjSzRvs [][]termRvs

/* build the second layer reverse list */
func conjReverse2(conj *Conj) {
	conjSzRvsLock.Lock()
	defer conjSzRvsLock.Unlock()

	if conj.size >= len(conjSzRvs) {
		resizeConjSzRvs(conj.size + 1)
	}

	termRvsList := conjSzRvs[conj.size]
	defer func() { conjSzRvs[conj.size] = termRvsList }()

	if termRvsList == nil {
		termRvsList = make([]termRvs, 0)
	}

	amts_.mutex.RLock()
	defer amts_.mutex.RUnlock()

	for _, amtId := range conj.amts {
		termRvsList = insertTermRvsList(conj.id, amtId, termRvsList)
	}
	if conj.size == 0 {
		insertTermRvsListEmptySet(conj.id)
	}
}

func insertTermRvsListEmptySet(cid int) {
	termrvslist := conjSzRvs[0]
	clist := termrvslist[0].cList
	defer func() { termrvslist[0].cList = clist }()
	clist = insertClist(cid, true, clist)
}

func resizeConjSzRvs(size int) {
	ASSERT(size >= len(conjSzRvs))
	size = upperPowerOfTwo(size)
	tmp := make([][]termRvs, size)
	copy(tmp[:len(conjSzRvs)], conjSzRvs[:])
	conjSzRvs = tmp
	ASSERT(len(conjSzRvs) == size)
}

func upperPowerOfTwo(size int) int {
	a := 4
	for a < size && a > 1 {
		a = a << 1
	}
	ASSERT(a > 1) /* avoid overflow */
	return a
}

func insertTermRvsList(conjId int, amtId int, list []termRvs) []termRvs {
	amt := &amts_.amts[amtId]

	for _, tid := range amt.terms {
		idx := sort.Search(len(list), func(i int) bool { return list[i].termId >= tid })
		if idx < len(list) && list[idx].termId == tid {
			/* term found */
			clist := list[idx].cList
			if clist == nil {
				clist = make([]cPair, 0)
			}
			clist = insertClist(conjId, amt.belong, clist)
			list[idx].cList = clist
		} else {
			/* term has not been found */
			clist := make([]cPair, 0, 1)
			clist = append(clist, cPair{conjId: conjId, belong: amt.belong})
			list = append(list, termRvs{termId: tid, cList: clist})
			n := len(list)
			if n > 1 && list[n-1].termId < list[n-2].termId {
				/* sort this list */
				sort.Sort(termRvsSlice(list))
			}
		}
	}
	return list
}

func insertClist(conjId int, belong bool, l []cPair) []cPair {
	idx := sort.Search(len(l), func(i int) bool {
		if l[i].conjId == conjId {
			return !l[i].belong || l[i].belong == belong
		}
		return l[i].conjId >= conjId
	})
	if idx < len(l) && (l[idx].conjId == conjId && l[idx].belong == belong) {
		/* found */
		return l
	}
	l = append(l, cPair{conjId: conjId, belong: belong})
	n := len(l)
	if n > 1 && !cPairSlice(l).Less(n-2, n-1) {
		sort.Sort(cPairSlice(l))
	}
	return l
}
