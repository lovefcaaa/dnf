package main

import (
	"commitor"
	"dnf"
	"service"

	"time"
)

func main() {
	commitor.Init()
	dnf.Init()
	go commitor.CommitLoop()

	time.Sleep(3 * time.Second)
	dnf.DisplayDocs()

	service.HttpServe("/ad/search", "/ad/zone", 7777)
}
