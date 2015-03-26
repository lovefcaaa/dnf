package main

import (
	"commitor"
	"dnf"
	"runtime"
	"service"

	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	commitor.Init()
	dnf.Init()
	go commitor.CommitLoop()

	time.Sleep(2 * time.Second)
	//dnf.DisplayDocs()

	go service.TcpServe()
	service.HttpServe("/ad/search", "/ad/zone", 7777)
}
