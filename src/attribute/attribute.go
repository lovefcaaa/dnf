package attribute

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"
)

type Attr struct {
	Adid            string /* 广告id */
	DnfDesc         string /* DNF描述 */
	Duration        int    /* 广告时长 */
	CreativeType    string /* 广告素材类型 */
	Adurl           string /* 广告图片地址 */
	Landing         string /* 广告落地页 */
	Width           string /* 广告图片宽 */
	Height          string /* 广告图片高 */
	Interval        int    /* 音频广告展示间隔 */
	SubTitle        string /* 广告下方文字 */
	Skin            string /* 推荐位的点进去播放页的皮肤图片 */
	SplashLanding   string /* 开屏点进去的推荐节目，格式/[catId]/[chanId]/[proId]/[0|1] */
	InternalLanding string /* banner内链跳转地址，格式/[catId]/[chanId]/[proId]/[0|1] */
	Tr              TimeRange
	Trackers        []Tracker
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
		"Skin: " + attr.Skin + "\n" +
		"SplashLanding" + attr.SplashLanding + "\n" +
		"InternalLanding" + attr.InternalLanding + "\n" +
		"TimeRange: " + attr.Tr.ToString() + "\n" +
		"Trackers: " + trackers + "\n"
}

const (
	GT = iota
	LT
)

type timePoint struct {
	t    int
	flag int // GT or LT
}

type TimeRange struct {
	tp  []timePoint
	gts int
	lts int
}

func (tr *TimeRange) Init() {
	tr.tp = make([]timePoint, 0, 2)
	tr.gts = 0
	tr.lts = 0
}

func (tr *TimeRange) ToString() (s string) {
	status := "[correct] "
	for i := 0; i != tr.Len(); i++ {
		var op string
		if tr.tp[i].flag == GT {
			op = ">="
		} else {
			op = "<="
		}
		s += fmt.Sprint(op, tr.tp[i].t, " ")
		if i+1 != tr.Len() && tr.tp[i].flag == tr.tp[i+1].flag {
			status = "[error] "
		}
	}
	s = status + s
	return
}

func (tr *TimeRange) Len() int {
	return len(tr.tp)
}

func (tr *TimeRange) Less(i, j int) bool {
	if tr.tp[i].t == tr.tp[j].t {
		return tr.tp[i].flag < tr.tp[j].flag
	}
	return tr.tp[i].t < tr.tp[j].t
}

func (tr *TimeRange) Swap(i, j int) {
	swap := func(a, b *int) {
		*a ^= *b
		*b ^= *a
		*a ^= *b
	}
	swap(&tr.tp[i].t, &tr.tp[j].t)
	swap(&tr.tp[i].flag, &tr.tp[j].flag)
}

func (tr *TimeRange) AddStart(start int) {
	tr.tp = append(tr.tp, timePoint{t: start, flag: GT})
	tr.gts++
	sort.Sort(tr)
}

func (tr *TimeRange) AddEnd(end int) {
	tr.tp = append(tr.tp, timePoint{t: end, flag: LT})
	tr.lts++
	sort.Sort(tr)
}

func (tr *TimeRange) CoverToday() (bool, error) {
	if tr.Len() == 0 {
		/* Fast path */
		return true, nil
	}
	now := time.Now()
	return tr.CoverTime(now.Year()*10000 + int(now.Month())*100 + now.Day())
}

func (tr *TimeRange) in(i int, day int) bool {
	switch {
	case i == tr.Len():
		return tr.tp[i-1].flag == GT
	case i == 0:
		return tr.tp[0].flag == LT
	default:
		/* day >= tr.tp[i-1].t && day <= tr.tp[i].t */
		return tr.tp[i-1].flag == GT && tr.tp[i].flag == LT
	}
}

func (tr *TimeRange) CoverTime(day int) (bool, error) {
	if tr.Len() == 0 {
		return true, nil
	}
	delta := tr.lts - tr.gts
	if delta*delta > 1 {
		/* delta should be equal to -1,0,1 */
		return false, errors.New("TimeRange delta illegal: " + strconv.Itoa(delta))
	}
	for i := 1; i < len(tr.tp); i++ {
		if tr.tp[i-1].flag == tr.tp[i].flag {
			return false, errors.New(fmt.Sprint("TimeRange error in ", tr.tp[i-1]))
		}
	}
	i := sort.Search(tr.Len(), func(i int) bool {
		if day == tr.tp[i].t {
			return tr.tp[i].flag == LT
		}
		return day < tr.tp[i].t
	})
	return tr.in(i, day), nil
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
