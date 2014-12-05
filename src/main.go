package main

import (
	"dnf"

	"fmt"
	"strconv"
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

func main() {
	dnf.Init()

	addDocs()

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
}
