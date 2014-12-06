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

	terms_.RLock()
	defer terms_.RUnlock()

	var key, op string

	if this.belong {
		op = "∈"
	} else {
		op = "∉"
	}
	key = terms_.terms[this.terms[0]].key
	s := fmt.Sprintf("%s %s { ", key, op)
	for i, idx := range this.terms {
		s += terms_.terms[idx].val
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
	amts_.RLock()
	defer amts_.RUnlock()
	s := "( "
	for i, idx := range this.amts {
		s += amts_.amts[idx].ToString()
		if i+1 < len(this.amts) {
			s += " ∩ "
		}
	}
	return s + " )"
}

func (this *Doc) ToString() (s string) {
	if len(this.conjs) == 0 {
		return ""
	}
	conjs_.RLock()
	defer conjs_.RUnlock()
	for i, idx := range this.conjs {
		s += conjs_.conjs[idx].ToString()
		if i+1 < len(this.conjs) {
			s += " ∪ "
		}
	}
	return
}

func (this *docList) display() {
	this.RLock()
	defer this.RUnlock()
	for i, doc := range this.docs {
		fmt.Println("Doc[", i, "]:", doc.ToString())
	}
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
	display(docs_)
}

func DisplayConjs() {
	display(conjs_)
}

func DisplayAmts() {
	display(amts_)
}

func DisplayTerms() {
	display(terms_)
}

func DisplayConjRevs() {
	fmt.Println("reverse list 1: ")
	conjRvsLock.RLock()
	defer conjRvsLock.RUnlock()
	for i, docs := range conjRvs {
		s := fmt.Sprint("conj[", i, "]: -> ")
		for _, id := range docs {
			s += strconv.Itoa(id) + " -> "
		}
		fmt.Println(s)
	}
}

func DisplayConjRevs2() {
	fmt.Println("reverse list 2: ")

	conjSzRvsLock.RLock()
	defer conjSzRvsLock.RUnlock()

	terms_.RLock()
	defer terms_.RUnlock()

	for i := 0; i < len(conjSzRvs); i++ {
		termlist := conjSzRvs[i]
		if termlist == nil || len(termlist) == 0 {
			continue
		}
		fmt.Println("***** size:", i, "*****")
		for _, termrvs := range termlist {
			s := fmt.Sprint(terms_.terms[termrvs.termId].ToString(), " -> ")
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
