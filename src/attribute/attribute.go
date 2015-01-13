package attribute

import (
	"strconv"
	"time"
)

type Attr struct {
	Adid         string /* 广告id */
	DnfDesc      string /* DNF描述 */
	Duration     int    /* 广告时长 */
	CreativeType string /* 广告素材类型 */
	Adurl        string /* 广告图片地址 */
	Landing      string /* 广告落地页 */
	Width        string /* 广告图片宽 */
	Height       string /* 广告图片高 */
	Interval     int    /* 音频广告展示间隔 */
	SubTitle     string /* 广告下方文字 */
	Tr           TimeRange
	Trackers     []Tracker
}

func (attr *Attr) ToString() string {
	var trackers string
	for i := 0; i != len(attr.Trackers); i++ {
		trackers += attr.Trackers[i].ToString()
	}
	return "Adid: " + attr.Adid + "\n" +
		"DnfDesc: " + attr.DnfDesc + "\n" +
		"Duration: " + strconv.Itoa(attr.Duration) + "\n" +
		"CreativeType: " + attr.CreativeType + "\n" +
		"Adurl: " + attr.Adurl + "\n" +
		"Landing: " + attr.Landing + "\n" +
		"Width: " + attr.Width + "\n" +
		"Height: " + attr.Height + "\n" +
		"Interval: " + strconv.Itoa(attr.Interval) + "\n" +
		"SubTitle: " + attr.SubTitle + "\n" +
		"TimeRange: " + attr.Tr.ToString() + "\n" +
		"Trackers: " + trackers + "\n"
}

type TimeRange struct {
	Startday int
	Endday   int
}

func (tr *TimeRange) CoverToday() bool {
	now := time.Now()
	today := now.Year()*10000 +
		int(now.Month())*100 +
		now.Day()
	return today >= tr.Startday && today <= tr.Endday
}

func (tr *TimeRange) ToString() string {
	return "{" + strconv.Itoa(tr.Startday) + ", " + strconv.Itoa(tr.Endday) + "}"
}

type Tracker struct {
	Event_type string
	Provider   string
	Url        string
}

func (tracker *Tracker) ToString() string {
	return "{ event_type: " + tracker.Event_type +
		", provider: " + tracker.Provider +
		", url:" + tracker.Url + " }"
}
