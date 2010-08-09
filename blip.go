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
	defaultSleeptime = 60 * 1000000000
	monitorDevice    = "/dev/tty.usbserial"
	connString       = "pgsql://localhost:54321/foo"
)

type blip int64

func spawnFetcher() chan blip {
	c := make(chan blip, inFlightMessages)
	go func() {
		sp, err := termios.Open(monitorDevice, os.O_RDONLY, 0)
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

func storeInDb(l *list.List) bool {
	/*
	conn, err := db.Connect(connString)
	if err != nil {
		syslog.Log(err)
		return false
	}
	defer conn.Close()
	ps = conn.Prepare("INSERT INTO blip (tsamp) VALUES ?")
	for item := range l.Iter() {
		ps.Execute(item)
	}
	 */
	return true
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
			if storeInDb(l) {
				// Nothing went wrong in the store, reset the in-flight list
				l = list.New()
			}
			time.Sleep(defaultSleeptime)
		}
	}
}
