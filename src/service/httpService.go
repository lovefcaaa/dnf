package service

import (
	"dnf"

	"fmt"
	"net/http"
	"strconv"
)

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "only get support", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "parse form error", http.StatusForbidden)
		return
	}
	conds := make([]dnf.Cond, 0)
	for k, v := range r.Form {
		if len(v) != 1 {
			continue
		}
		conds = append(conds, dnf.Cond{Key: k, Val: v[0]})
	}

	var rc string
	if docs, err := dnf.Search(conds); err != nil {
		rc = fmt.Sprint("dnf search err: ", err)
		fmt.Println("dnf search err:", err)
	} else if len(docs) == 0 {
		rc = "NULL"
	} else {
		for _, doc := range docs {
			rc += fmt.Sprint("[", doc, "]->")
		}
		fmt.Println("search result:", rc)
	}
	http.Error(w, rc, http.StatusOK)
}

func HttpServe(searchUrl string, port int) {
	http.HandleFunc(searchUrl, searchHandler)
	panic(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
