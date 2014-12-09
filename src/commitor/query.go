package commitor

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"attribute"
	_ "github.com/ziutek/mymysql/godrv"
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
		fmt.Printf("attr: %+v", attr)
	}
}

//=======================================================

type rowsClosure func(rows *sql.Rows) interface{}

func dbQuery(f rowsClosure, query string, args ...interface{}) (rc interface{}, err error) {
	var rows *sql.Rows
	if rows, err = db.Query(query, args...); err == nil {
		rc = f(rows)
	}
	return
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

func trackerUnmarshal(tracker string) ([]attribute.Tracker, error) {
	var trackers []attribute.Tracker
	//var trackerObj attribute.Tracker
	dec := json.NewDecoder(strings.NewReader(tracker))
	if err := dec.Decode(&trackers); err != nil {
		return nil, err
	}
	fmt.Println("unmarshal len :", len(trackers))
	return trackers, nil
}

func getAdAttr(adid string) (attrs attribute.Attr, err error) {
	f := func(rows *sql.Rows) interface{} {
		if !rows.Next() {
			return nil
		}

		var adid, adurl, landing, tracker, width, height string

		err := rows.Scan(&adid, &adurl,
			&landing, &tracker, &width, &height)

		if err != nil {
			fmt.Println("in get ad attr, scan rows err:", err)
			return nil
		}

		var trackers []attribute.Tracker
		if trackers, err = trackerUnmarshal(tracker); err != nil {
			fmt.Println("in get ad attr, trackers unmarshal err:", err)
			return nil
		}

		return attribute.Attr{
			Adid:     adid,
			Adurl:    adurl,
			Landing:  landing,
			Width:    width,
			Height:   height,
			Trackers: trackers,
		}
	}

	query := "SELECT bannerid, imageurl, url, tracker, width, height from qtad_banners where bannerid = ?"

	var rc interface{}
	rc, err = dbQuery(f, query, adid)
	attrs, _ = rc.(attribute.Attr)
	return
}
