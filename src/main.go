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

	time.Sleep(1 * time.Second)
	dnf.DisplayDocs()

	service.HttpServe("/ad/search", 7777)
}
