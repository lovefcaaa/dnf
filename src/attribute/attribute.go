package attribute

import ()

type Attr struct {
	Adid     string /* 广告id */
	Duration int    /* 广告时长 */
	Adurl    string /* 广告图片地址 */
	Landing  string /* 广告落地页 */
	Width    string /* 广告图片宽 */
	Height   string /* 广告图片高 */
	Trackers []Tracker
}

type Tracker struct {
	Event_type string
	Provider   string
	Url        string
}
