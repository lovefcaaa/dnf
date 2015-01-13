package service

import (
	"commitor"
	"dnf"

	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func httpZonesHandler(w http.ResponseWriter, r *http.Request) {
	var version int
	if r.Method != "GET" {
		http.Error(w, "only get support", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "parse form error", http.StatusForbidden)
		return
	}
	if len(r.Form["version"]) != 0 {
		version, _ = strconv.Atoi(r.Form["version"][0])
	}
	h := commitor.GetZonesInfoHandler()
	infos := h.GetZonesInfo(version)
	rcMap := make(map[string][]commitor.ZoneInfo)
	rcMap["zones"] = infos
	if rc, err := json.Marshal(rcMap); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		http.Error(w, string(rc), http.StatusOK)
	}
}

func httpSearchHandler(w http.ResponseWriter, r *http.Request) {
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
	h := dnf.GetHandler()
	if h == nil {
		http.Error(w, "interal error", http.StatusOK)
		return
	}

	if docs, err := h.Search(conds); err != nil {
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

func HttpServe(searchUrl string, zoneUrl string, port int) {
	http.HandleFunc(zoneUrl, httpZonesHandler)
	http.HandleFunc(searchUrl, httpSearchHandler)
	panic(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
