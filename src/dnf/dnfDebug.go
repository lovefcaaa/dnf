package dnf

import (
	"fmt"
	"strconv"
)

var debug bool = true

var DEBUG func(msg ...interface{})
var ASSERT func(expression bool)

func doDEBUG(msg ...interface{}) {
	if debug {
		fmt.Println(msg)
	}
}

func noDEBUG(msg ...interface{}) {}

func doASSERT(expression bool) {
	if debug && !(expression) {
		panic("ASSERT")
	}
}

func noASSERT(expression bool) {}

func debugInit() {
	if debug {
		DEBUG = doDEBUG
		ASSERT = doASSERT
	} else {
		DEBUG = noDEBUG
		ASSERT = noASSERT
	}
}

func (this *Term) ToString() string {
	if this.id == 0 {
		/* empty set */
		return " ∅ "
	}
	return fmt.Sprintf("( %s  %s )", this.key, this.val)
}

func (this *Amt) ToString() string {
	if len(this.terms) == 0 {
		return ""
	}

	h := GetHandler()
	h.terms_.RLock()
	defer h.terms_.RUnlock()

	var key, op string

	if this.belong {
		op = "∈"
	} else {
		op = "∉"
	}
	key = h.terms_.terms[this.terms[0]].key
	s := fmt.Sprintf("%s %s { ", key, op)
	for i, idx := range this.terms {
		s += h.terms_.terms[idx].val
		if i+1 < len(this.terms) {
			s += ", "
		}
	}
	return s + " }"
}

func (this *Conj) ToString() string {
	if len(this.amts) == 0 {
		return ""
	}
	/* bugs to fix here */
	h := GetHandler()
	h.amts_.RLock()
	defer h.amts_.RUnlock()
	s := "( "
	for i, idx := range this.amts {
		s += h.amts_.amts[idx].ToString()
		if i+1 < len(this.amts) {
			s += " ∩ "
		}
	}
	return s + " )"
}

func (this *Doc) ToString() (s string) {
	if len(this.conjs) == 0 {
		s = "len(conjs == 0)"
	}
	h := GetHandler()
	h.conjs_.RLock()
	defer h.conjs_.RUnlock()
	for i, idx := range this.conjs {
		s += h.conjs_.conjs[idx].ToString()
		if i+1 < len(this.conjs) {
			s += " ∪ "
		}
	}
	s += "\n"
	s += this.attr.ToString()
	return
}

func (this *docList) display() {
	this.RLock()
	defer this.RUnlock()
	println("len docs == ", len(this.docs))
	for i, doc := range this.docs {
		fmt.Println("Doc[", i, "]:", doc.ToString())
	}
}

func (this *docList) docId2Map(docid int) map[string]interface{} {
	if len(this.docs) <= docid {
		return nil
	}
	this.RLock()
	defer this.RUnlock()
	m := make(map[string]interface{})
	doc := &this.docs[docid]
	m["id"] = doc.docid

	if doc.attr.CreativeType == "banner" {
		m["creativetype"] = 0
		m["image"] = doc.attr.Adurl
		m["landing"] = doc.attr.Landing
		w, _ := strconv.Atoi(doc.attr.Width)
		h, _ := strconv.Atoi(doc.attr.Height)
		m["size"] = w*10000 + h
		m["audio"] = ""
		m["duration"] = 0
	} else {
		m["creativetype"] = 1
		m["image"] = ""
		m["landing"] = ""
		m["size"] = 0
		m["audio"] = doc.attr.Adurl
		m["duration"] = doc.attr.Duration
	}
	trackerSlice := make([]map[string]string, 0, 1)
	for _, tr := range doc.attr.Trackers {
		t := make(map[string]string)
		if len(tr.Event_type) != 0 {
			t["event_type"] = tr.Event_type
		}
		if len(tr.Provider) != 0 {
			t["provider"] = tr.Provider
		}
		if len(tr.Url) != 0 {
			t["url"] = tr.Url
		}
		trackerSlice = append(trackerSlice, t)
	}
	m["trackers"] = trackerSlice

	return m
}

func DocId2Map(docid int) map[string]interface{} {
	return GetHandler().docs_.docId2Map(docid)
}

func (this *conjList) display() {
	this.RLock()
	defer this.RUnlock()
	for i, conj := range this.conjs {
		fmt.Println("Conj[", i, "]", "size:", conj.size, conj.ToString())
	}
}

func (this *amtList) display() {
	this.RLock()
	defer this.RUnlock()
	for i, amt := range this.amts {
		fmt.Println("Amt[", i, "]:", amt.ToString())
	}
}

func (this *termList) display() {
	this.RLock()
	defer this.RUnlock()
	for i, term := range this.terms {
		fmt.Println("Term[", i, "]", term.ToString())
	}
}

type displayer interface {
	display()
}

func display(obj displayer) {
	obj.display()
}

func DisplayDocs() {
	display(GetHandler().docs_)
}

func DisplayConjs() {
	display(GetHandler().conjs_)
}

func DisplayAmts() {
	display(GetHandler().amts_)
}

func DisplayTerms() {
	display(GetHandler().terms_)
}

func DisplayConjRevs() {
	fmt.Println("reverse list 1: ")
	h := GetHandler()
	h.conjRvsLock.RLock()
	defer h.conjRvsLock.RUnlock()
	for i, docs := range h.conjRvs {
		s := fmt.Sprint("conj[", i, "]: -> ")
		for _, id := range docs {
			s += strconv.Itoa(id) + " -> "
		}
		fmt.Println(s)
	}
}

func DisplayConjRevs2() {
	fmt.Println("reverse list 2: ")

	h := GetHandler()
	h.conjSzRvsLock.RLock()
	defer h.conjSzRvsLock.RUnlock()

	h.terms_.RLock()
	defer h.terms_.RUnlock()

	for i := 0; i < len(h.conjSzRvs); i++ {
		termlist := h.conjSzRvs[i]
		if termlist == nil || len(termlist) == 0 {
			continue
		}
		fmt.Println("***** size:", i, "*****")
		for _, termrvs := range termlist {
			s := fmt.Sprint(h.terms_.terms[termrvs.termId].ToString(), " -> ")
			for _, cpair := range termrvs.cList {
				var op string
				if cpair.belong {
					op = "∈"
				} else {
					op = "∉"
				}
				s += fmt.Sprintf("(%d %s) -> ", cpair.conjId, op)
			}
			fmt.Println("   ", s)
		}
	}
}
