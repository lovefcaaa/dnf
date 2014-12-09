package attribute

import ()

type Attr struct {
	adid     string
	duration int
	adurl    string
	trackers []Tracker
}

type Event int

const (
	DISPLAY = iota
	CLICK
)

type Tracker struct {
	event    Event
	provider string
	url      string
}
