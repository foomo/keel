package main

import (
	"time"
)

func main() {
	go server()

	go client()

	time.Sleep(time.Second * 60)
}
