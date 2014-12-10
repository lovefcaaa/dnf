package commitor

/*
临时使用的测试文件
*/

import (
	"fmt"
)

// for test
func GetAssocAdZone(adid string) {
	if ids, err := getAssocAdZone(adid); err != nil {
		fmt.Println("get assoc ad zone err: ", err)
	} else {
		fmt.Println("get adid: ", adid, " slice: ", ids)
	}
}

func GetAdAttr(adid string) {
	attr, err := getAdAttr(adid)
	if err != nil {
		fmt.Println("GetAdAttr err: ", err)
	} else {
		fmt.Printf(attr.ToString())
	}
}

func PrintAdAttr(adid string) {
	doc := ad2Doc(adid)
	if doc != nil {
		fmt.Println(doc.GetAttr())
	}
}
