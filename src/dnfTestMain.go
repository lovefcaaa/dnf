package main

import (
	"dnf"

	"fmt"
	"strconv"
	"sync"
)

var docs []string = []string{
	"(age in {3} and state in {NY}) or (state in {CA} and gender in {M})",
	"(age in {3} and gender in {F}) or (state not in {CA, NY})",
	"(age in {3} and gender in {M} and state not in {CA}) or (state in {CA} and gender in {F})",
	"(age in {3, 4}) or (state in {CA} and gender in {M})",
	"(state not in {CA, NY}) or (age in {3, 4})",
	"(state not in {CA, NY}) or (age in {3} and state in {NY}) or (state in {CA} and gender in {M})",
	"(age in {3} and state in {NY}) or (state in {CA} and gender in {F})",
}

func check() {
	for i, doc := range docs {
		fmt.Println()
		if err := dnf.DnfCheck(doc); err != nil {
			fmt.Println("***FATAL***:", err, " doc[", i, "]: ", doc)
		} else {
			fmt.Println("ACCEPT: doc[", i, "]:", doc)
		}
	}
}

func addDocs() {
	for i, doc := range docs {
		if err := dnf.AddDoc("doc"+strconv.Itoa(i), strconv.Itoa(i), doc); err != nil {
			fmt.Println("add doc[", strconv.Itoa(i), "] error:", err)
			return
		}
	}
}

func addDocsRace() {
	task := func(i int, doc string, wg *sync.WaitGroup) {
		if err := dnf.AddDoc("doc"+strconv.Itoa(i), strconv.Itoa(i), doc); err != nil {
			fmt.Println("add doc[", strconv.Itoa(i), "] error:", err)
			panic(err)
		}
		wg.Done()
	}
	var wg sync.WaitGroup
	for i, doc := range docs {
		wg.Add(1)
		go task(i, doc, &wg)
	}
	wg.Wait()
}

func dnf_test_main() {
	//dnf.Init()
	//addDocsRace()

	dnf.DisplayDocs()

	fmt.Println()
	dnf.DisplayConjs()

	fmt.Println()
	dnf.DisplayAmts()

	fmt.Println()
	dnf.DisplayTerms()

	fmt.Println()
	dnf.DisplayConjRevs()

	fmt.Println()
	dnf.DisplayConjRevs2()

	fmt.Println()
	//conds := []dnf.Cond{{"zone", "1"}}
	//conds := []dnf.Cond{{"state", "CA"}, {"age", "4"}}
	//conds := []dnf.Cond{{"gender", "M"}}
	// resultDocs, err := dnf.Search(conds)
	// if err != nil {
	// 	fmt.Println("search error: ", err)
	// 	return
	// }
	// s := ""
	// for _, docid := range resultDocs {
	// 	s += strconv.Itoa(docid) + " -> "
	// }
	// fmt.Println("conds: ", conds)
	// fmt.Println("found doc: ", s)
}

func commitor_test_main() {
	commitor.Init()
	fmt.Println("========= 1 =============")
	commitor.PrintAdAttr("1")
	fmt.Println("========= 2 =============")
	commitor.PrintAdAttr("2")
	fmt.Println("========= 3 =============")
	commitor.PrintAdAttr("3")
}

func commitor_and_dnf() {
	commitor.Init()
	dnf.Init()
	go commitor.CommitLoop()

	commitor.PrintAdAttr("1")
	commitor.PrintAdAttr("3")

	time.Sleep(1 * time.Second)

	dnf_test_main()

	fmt.Println("dnf_test_main end")

	conds := []dnf.Cond{{"zone", "1"}, {"region", "07"}, {"width", "200"}, {"height", "200"}}
	resultDocs, err := dnf.Search(conds)
	if err != nil {
		fmt.Println("search error: ", err)
		return
	}
	s := ""
	for _, docid := range resultDocs {
		s += strconv.Itoa(docid) + " -> "
	}
	fmt.Println("conds: ", conds)
	fmt.Println("found doc: ", s)
}
