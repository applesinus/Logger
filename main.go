package main

import (
	"logger/opt1"
	"time"
)

func main() {
	l := opt1.NewLogger(time.Second*2, 1, 4)
	l.Log(1, "Hello")
	l.Log(2, "World")
	l.Log(3, "!!!")
	l.Log(2, "Kek")
	l.Log(2, "Lol")
	l.Log(2, "Prekol")

	time.Sleep(time.Second * 10)
}
