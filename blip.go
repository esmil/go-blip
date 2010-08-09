package main

import (
	"fmt"
	"container/list"
	"os"
	"termios"
	"time"
)

const (
	version          = "v0.1" // Version of program
	inFlightMessages = 300    // How many blips do we allow to be in-flight?
)

type blip int64

func spawnFetcher() chan blip {
	c := make(chan blip, inFlightMessages)
	go func() {
		sp, err := termios.Open("/dev/tty.usbserial", os.O_RDONLY, 0)
		if err != nil {
			panic(err)
		}
		//sp.Set(termios.B9600)

		b := make([]byte, 1)
		for {
			sp.Read(b)
			c <- blip(time.Nanoseconds())
		}
	}()

	return c
}

func storeInDb(l *list.List) {
	return
}

func main() {
	fmt.Printf("Blip storage daemon %s\n", version)
	fetchC := spawnFetcher()
	l := list.New()

	for {
		select {
		case x := <-fetchC:
			l.PushBack(x)
			fmt.Printf("Read blip %s", x)
		default:
			storeInDb(l)
			time.Sleep(60 * 1000000000)
		}
	}
}
