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
		defer rows.Close()
		rc = f(rows)
	}
	return
}

//TODO: 这个逻辑放在这个包里面不合理

type ZoneInfo struct {
	Zoneid   string
	Width    string
	Height   string
	Version  string
	Comments string
}

type ZonesInfoHandler struct {
	m map[int][]ZoneInfo
}

var zonesInfoHandler *ZonesInfoHandler

func GetZonesInfoHandler() *ZonesInfoHandler {
	return zonesInfoHandler
}

func (h *ZonesInfoHandler) GetZonesInfo(version int) []ZoneInfo {
	return h.m[version]
}

func zoneInfoCommit() {
	h := &ZonesInfoHandler{
		m: make(map[int][]ZoneInfo),
	}
	defer func() { zonesInfoHandler = h }()
	h.m[5] = getZonesInfo(5)
	h.m[6] = getZonesInfo(6)
}

func getZonesInfo(version int) (rc []ZoneInfo) {
	f := func(rows *sql.Rows) interface{} {
		info := make([]ZoneInfo, 0)
		for rows.Next() {
			var zoneid, width, height, comments string
			if err := rows.Scan(&zoneid, &width, &height, &comments); err != nil {
				fmt.Println("scan zone err:", err)
				continue
			}
			info = append(info, ZoneInfo{
				Zoneid:   zoneid,
				Width:    width,
				Height:   height,
				Comments: comments,
			})
		}
		return info
	}
	zoneInfoInter, err := dbQuery(f, "SELECT zoneid, width, height, comments from qtad_zones where version = ?", version)
	if err != nil {
		fmt.Println("GetZonesInfo db query error: ", err)
		return
	}
	rc, _ = zoneInfoInter.([]ZoneInfo)
	return
}

func adCommit() {
	f := func(rows *sql.Rows) interface{} {
		banners := make([]string, 0)
		var adid string
		for rows.Next() {
			if err := rows.Scan(&adid); err != nil {
				fmt.Println("scan banner err:", err)
				continue
			}
			banners = append(banners, adid)
		}
		return banners
	}
	idsInter, err := dbQuery(f, "SELECT bannerid from qtad_banners")
	if err != nil {
		fmt.Println("adCommit db query error:", err)
		return
	}
	if ids, ok := idsInter.([]string); !ok {
		fmt.Println("adCommit interface query error")
		return
	} else {
		h := dnf.NewHandler()
		for _, id := range ids {
			if doc := ad2Doc(id); doc != nil {
				err := h.AddDoc(doc.GetName(), doc.GetDocId(), doc.GetDnf(), doc.GetAttr())
				if err != nil {
					fmt.Println("adCommit add doc err: ", err,
						"id: ", doc.GetDocId(),
						"dnf: ", doc.GetDnf())
				} else {
					// fmt.Println("ad doc [", id, "] ok")
				}
			}
		}
		dnf.SaveHandler(h)
	}
}

func ad2Doc(adid string) *dnf.Doc {
	attr, err := getAdAttr(adid)
	if err != nil {
		fmt.Println("getAdAttr error: ", err)
		return nil
	}
	if ok, _ := attr.Tr.CoverToday(); !ok {
		return nil
	}
	zones, err2 := getAssocAdZone(adid)
	if err2 != nil {
		fmt.Println("getAssocAdZone error: ", err2)
		return nil
	}
	zoneDnf, err3 := zones2Dnf(zones)
	if err3 != nil {
		fmt.Println("zones2Dnf error: ", err3)
		return nil
	}
	if len(zoneDnf) != 0 {
		if len(attr.DnfDesc) != 0 {
			attr.DnfDesc += " and "
		}
		attr.DnfDesc += zoneDnf
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

func parseValue(param string) string {
	arr := strings.SplitN(param, ":", 3)
	if len(arr) != 3 {
		fmt.Println("parseValue: param: ", param, "format error")
		return ""
	}
	val := strings.Trim(arr[2], " \n\"")
	size, _ := strconv.Atoi(arr[1])
	if len(val) == size {
		return val
	}
	fmt.Println(val, "parseValue len error: len(val): ", len(val), "advise length: ", size)
	return val
}

func parseParameters(param string) (m map[string]string, ok bool) {
	arr := strings.SplitN(param, ":", 3)
	if len(arr) != 3 {
		return nil, false
	}
	param = strings.Trim(arr[2], "{} \n")
	arr = strings.Split(param, ";")
	m = make(map[string]string)

	for i, s := range arr {
		switch {
		case s == "s:19:\"vast_video_duration\"":
			m["duration"] = parseValue(arr[i+1])
		case s == "s:15:\"vast_video_type\"":
			m["type"] = parseValue(arr[i+1])
		case s == "s:28:\"vast_video_outgoing_filename\"":
			m["outgoing"] = parseValue(arr[i+1])
		case s == "s:27:\"vast_video_clickthrough_url\"":
			m["landing"] = parseValue(arr[i+1])
		}
	}

	return m, true
}

func byteInterToString(input interface{}) (rc string) {
	if b, ok := input.([]byte); ok {
		rc = string(b)
	}
	return
}

func commentsParse(comment string) map[string]string {
	m := make(map[string]string)
	terms := strings.Split(comment, ";")
	for _, term := range terms {
		if kv := strings.Split(term, "="); len(kv) == 2 {
			m[kv[0]] = kv[1]
		}
	}
	return m
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

		var duration int
		var creativeType string = "banner" // default
		var adid, adurl, landing, tracker, width, height, comments, conds, parameters string

		var adidInter, adurlInter, landingInter, trackerInter, widthInter, heightInter, commentsInter, condsInter, parametersInter interface{} // deal with tracker = NULL in database

		err := rows.Scan(&adidInter, &adurlInter, &landingInter, &trackerInter,
			&widthInter, &heightInter, &commentsInter, &condsInter, &parametersInter)

		if err != nil {
			fmt.Println("in get ad attr, scan rows err:", err)
			return nil
		}

		adid = byteInterToString(adidInter)
		adurl = byteInterToString(adurlInter)
		landing = byteInterToString(landingInter)
		tracker = byteInterToString(trackerInter)
		width = byteInterToString(widthInter)
		height = byteInterToString(heightInter)
		comments = byteInterToString(commentsInter)
		conds = byteInterToString(condsInter)
		parameters = byteInterToString(parametersInter)

		/* width和height都为-3，说明是音频广告 */
		if width == "-3" && height == "-3" {
			if m, ok := parseParameters(parameters); ok {
				adurl = m["outgoing"]
				landing = m["landing"]
				duration, _ = strconv.Atoi(m["duration"])
				creativeType = m["type"]
			}
		}

		var dnf string
		var tr attribute.TimeRange

		if dnf, tr, err = parseDnfDesc(conds); err != nil {
			fmt.Println("limitation to dnf err: ", err)
			return nil
		}

		var trackers []attribute.Tracker

		if len(tracker) != 0 {
			if trackers, err = trackerUnmarshal(tracker); err != nil {
				fmt.Println("in get ad attr, trackers unmarshal err:", err,
					"tracker: ", tracker)
				return nil
			}
		}

		//音频广告，没有width和height参数(-3)
		if width != "-3" && height != "-3" {
			if len(dnf) != 0 {
				dnf += " and "
			}
			dnf += "width in {" + width + "} and height in {" + height + "}"
		}

		interval := 60 // default
		subtitle := ""
		skinLoading := ""
		splashLanding := ""
		internalLanding := ""
		kv := commentsParse(comments)

		/* --------- 在这里新增属性 --------- */
		if s, ok := kv["subtitle"]; ok {
			subtitle = s
		}
		if s, ok := kv["interval"]; ok {
			if n, err := strconv.Atoi(s); err == nil {
				interval = n
			}
		}
		if s, ok := kv["skinLoading"]; ok { /* 推荐位点进去的皮肤图片 */
			skinLoading = s
		}
		if s, ok := kv["splashLanding"]; ok {
			splashLanding = s
		}
		if s, ok := kv["internalLanding"]; ok {
			internalLanding = s
		}
		/* ---------------------------------- */

		return attribute.Attr{
			Adid:            adid,
			DnfDesc:         dnf,
			Duration:        duration,
			CreativeType:    creativeType,
			Adurl:           adurl,
			Landing:         landing,
			Width:           width,
			Height:          height,
			Interval:        interval,
			SubTitle:        subtitle,
			Skin:            skinLoading,
			SplashLanding:   splashLanding,
			InternalLanding: internalLanding,
			Tr:              tr,
			Trackers:        trackers,
		}
	}

	query := "SELECT bannerid, imageurl, url, tracker, width, height, comments, compiledlimitation, parameters from qtad_banners where bannerid = ?"

	var rc interface{}
	rc, err = dbQuery(f, query, adid)

	var ok bool
	if attr, ok = rc.(attribute.Attr); !ok {
		if err == nil {
			err = errors.New("db query adid error")
		}
	}
	return
}

func zones2Dnf(zones []string) (dnf string, err error) {
	if len(zones) == 0 {
		return "", nil
	}
	dnf = "zone in {"
	for i := 0; i != len(zones); i++ {
		/*
		   我们所使用的revive-adserver开源AE系统，
		   默认把所有广告都关联到0号广告位上
		*/
		dnf += zones[i]
		if i != len(zones)-1 {
			dnf += ", "
		}
	}
	dnf += " }"
	return
}

func trackerUnmarshal(tracker string) ([]attribute.Tracker, error) {
	var err error
	var trackers []attribute.Tracker

	dec := json.NewDecoder(strings.NewReader(tracker))
	if err = dec.Decode(&trackers); err != nil {
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

func parseOsCond(limit string) (amt string, err error) {
	// now, cond = MAX_checkClient_Os('android,ios', '=~')
	// or, cond = MAX_checkClient_Os('android,ios', '!~')

	cond := strings.Split(limit, "MAX_checkClient_Os")[1]
	// now, cond = ('android,ios', '=~')

	cond = strings.Map(runeCondMapping, cond)
	// now, cond = android,ios,=~

	tmp := strings.Split(cond, ",")
	n := len(tmp)
	if n < 2 {
		return "", errors.New("client format error: " + cond)
	}

	var op string
	switch tmp[n-1] {
	case "=~":
		op = " in "
	case "!~":
		op = " not in "
	default:
		return "", errors.New("unrecognize op: " + tmp[n-1])
	}

	amt = "phonetype" + op + "{" + tmp[0]
	for i := 1; i < n-1; i++ {
		amt += "," + tmp[i]
	}
	amt += "}"
	return amt, nil
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
	limits := strings.Split(conds, " and ")
	tr.Init()
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
			if start != 0 {
				tr.AddStart(start)
			}
			if end != 0 {
				tr.AddEnd(end)
			}

		case strings.Contains(limits[i], "MAX_checkClient_Os"):
			if amt, err = parseOsCond(limits[i]); err != nil {
				return dnf, tr, err
			}
			if len(dnf) != 0 {
				dnf += " and "
			}
			dnf += amt
		}
	}
	return
}
