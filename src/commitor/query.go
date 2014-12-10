package commitor

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"attribute"
	"dnf"
	_ "github.com/ziutek/mymysql/godrv"
)

type rowsClosure func(rows *sql.Rows) interface{}

func dbQuery(f rowsClosure, query string, args ...interface{}) (rc interface{}, err error) {
	var rows *sql.Rows
	if rows, err = db.Query(query, args...); err == nil {
		rc = f(rows)
	}
	return
}

func ad2Doc(adid string) *dnf.Doc {
	attr, err := getAdAttr(adid)
	if err != nil {
		fmt.Println("getAdAttr error: ", err)
		return nil
	}
	zones, err2 := getAssocAdZone(adid)
	if err2 != nil {
		fmt.Println("getAssocAdZone error: ", err2)
	}
	zoneDnf, err3 := zones2Dnf(zones)
	if err3 != nil {
		fmt.Println("zones2Dnf error: ", err3)
		return nil
	}
	if len(zoneDnf) != 0 {
		attr.DnfDesc += " and " + zoneDnf
	}
	return dnf.NewDoc(adid, "( "+attr.DnfDesc+" )", "", true, &attr)
}

/*
在我们的广告管理系统中，ad_id即bannerid
*/

/* 获取广告关联的广告位 */
func getAssocAdZone(adid string) (zoneids []string, err error) {
	f := func(rows *sql.Rows) interface{} {
		ids := make([]string, 0)
		for rows.Next() {
			var zoneid string
			if err := rows.Scan(&zoneid); err != nil {
				fmt.Println("scan zone_id err:", err)
			} else {
				ids = append(ids, zoneid)
			}
		}
		return ids
	}

	var rc interface{}
	rc, err = dbQuery(f, "SELECT zone_id from qtad_ad_zone_assoc where ad_id = ?", adid)
	zoneids, _ = rc.([]string)
	return
}

/*
   获取广告属性
   备注: 这里的属性只是从qtad_banners表中查询的，dnf不全，
         需要再调用getAssocAdZone去获取关联的广告位，补全dnf
*/
func getAdAttr(adid string) (attr attribute.Attr, err error) {
	f := func(rows *sql.Rows) interface{} {
		if !rows.Next() {
			return nil
		}

		var adid, adurl, landing, tracker, width, height, conds string

		err := rows.Scan(&adid, &adurl, &landing,
			&tracker, &width, &height, &conds)

		if err != nil {
			fmt.Println("in get ad attr, scan rows err:", err)
			return nil
		}

		var dnf string
		var tr attribute.TimeRange
		if dnf, tr, err = parseDnfDesc(conds); err != nil {
			fmt.Println("limitation to dnf err: ", err)
			return nil
		}

		var trackers []attribute.Tracker
		if trackers, err = trackerUnmarshal(tracker); err != nil {
			fmt.Println("in get ad attr, trackers unmarshal err:", err)
			return nil
		}

		dnf += " and width in {" + width + "} and height in {" + height + "}"

		return attribute.Attr{
			Adid:     adid,
			DnfDesc:  dnf,
			Adurl:    adurl,
			Landing:  landing,
			Width:    width,
			Height:   height,
			Tr:       tr,
			Trackers: trackers,
		}
	}

	query := "SELECT bannerid, imageurl, url, tracker, width, height, compiledlimitation from qtad_banners where bannerid = ?"

	var rc interface{}
	rc, err = dbQuery(f, query, adid)
	attr, _ = rc.(attribute.Attr)
	return
}

func zones2Dnf(zones []string) (dnf string, err error) {
	if len(zones) == 0 {
		return "", nil
	}
	dnf = "zone in {"
	for i := 0; i != len(zones); i++ {
		dnf += zones[i]
		if i != len(zones)-1 {
			dnf += ", "
		}
	}
	dnf += " }"
	return
}

func trackerUnmarshal(tracker string) ([]attribute.Tracker, error) {
	var trackers []attribute.Tracker
	dec := json.NewDecoder(strings.NewReader(tracker))
	if err := dec.Decode(&trackers); err != nil {
		return nil, err
	}
	return trackers, nil
}

func runeCondMapping(c rune) rune {
	switch {
	case c == '(' || c == ')' || c == '\'' || c == ' ':
		return rune(-1)
	}
	return c
}

func parseDateCond(limit string) (start int, end int) {
	// now, limit = MAX_checkTime_Date('20141211@Asia/Chongqing', '>=') only '>=' and '<=' supported

	cond := strings.Split(limit, "MAX_checkTime_Date")[1]
	// now, cond = ('20141211@Asia/Chongqing', '>=')

	cond = strings.Map(runeCondMapping, cond)
	// now, cond = 20141211@Asia/Chongqing,>=

	var date *int
	var datetime, op string
	var err error

	if tmp := strings.Split(cond, ","); len(tmp) != 2 {
		fmt.Println("time cond error: ", cond)
		return 0, 0
	} else {
		datetime = tmp[0]
		op = tmp[1]
	}
	// now, datetime = 20141211@Asia/Chongqing

	switch {
	case op == ">=":
		date = &start
	case op == "<=":
		date = &end
	default:
		fmt.Println("unsupport MAX_checkTime_Date op ", op)
		return 0, 0
	}

	if tmp := strings.Split(datetime, "@"); len(tmp) != 2 {
		fmt.Println("date format error: ", datetime)
		return 0, 0
	} else if *date, err = strconv.Atoi(tmp[0]); err != nil {
		fmt.Println("convert datetime to string error: ", err, "datetime: ", tmp[0])
	}

	return
}

func parseGeoCond(limit string) (amt string, err error) {
	// now, cond = MAX_checkGeo_Region('cn|01,02,03,06,07,32', '=~')

	cond := strings.Split(limit, "MAX_checkGeo_Region")[1]
	// now, cond = ('cn|01,02,03,06,07,32', '=~')

	cond = strings.Map(runeCondMapping, cond)
	// now, cond = cn|01,02,03,06,07,32,=~

	if tmp := strings.Split(cond, "|"); len(tmp) != 2 {
		return "", errors.New("geo format error, no char of |: " + cond)
	} else if tmp[0] != "cn" {
		return "", errors.New("unrecognize country: " + tmp[0])
	} else {
		cond = tmp[1]
	}
	// now, cond = 01,02,03,06,07,32,=~

	lastIdx := strings.LastIndex(cond, ",")

	var op string
	if string(cond[lastIdx+1:]) == "=~" {
		op = " in "
	} else if string(cond[lastIdx+1:]) == "!=" {
		op = " not in "
	} else {
		return "", errors.New("unrecognize op: " + string(cond[lastIdx+1:]))
	}

	amt = "region" + op + "{" + string(cond[:lastIdx]) + "}"

	return amt, nil
}

func parseDnfDesc(conds string) (dnf string, tr attribute.TimeRange, err error) {
	m := make(map[string]bool)
	limits := strings.Split(conds, "and")
	tr.Startday = 19900101
	tr.Endday = 29900101
	for i := 0; i < len(limits); i++ {
		var amt string

		switch {
		case strings.Contains(limits[i], "MAX_checkGeo_Region"):
			/* case Geo cond */
			if _, ok := m["MAX_checkGeo_Region"]; ok {
				return dnf, tr, errors.New("Geo Info Dup")
			}
			m["MAX_checkGeo_Region"] = true
			if amt, err = parseGeoCond(limits[i]); err != nil {
				return dnf, tr, err
			}
			if len(dnf) != 0 {
				dnf += " and "
			}
			dnf += amt

		case strings.Contains(limits[i], "MAX_checkTime_Date"):
			/* add more case here */
			start, end := parseDateCond(limits[i])
			if start > 0 && tr.Startday < start {
				tr.Startday = start
			}
			if end > 0 && tr.Endday > end {
				tr.Endday = end
			}
		}
	}
	return
}
