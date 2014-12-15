package attribute

import (
	"strconv"
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
		"TimeRange: " + attr.Tr.ToString() + "\n" +
		"Trackers: " + trackers + "\n"
}

type TimeRange struct {
	Startday int
	Endday   int
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
	return "{ Event_type: " + tracker.Event_type +
		", Provider: " + tracker.Provider +
		", Url:" + tracker.Url + " }"
}
