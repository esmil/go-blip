package main

import (
	"container/list"
	"fmt"
	"os"
	"pg"
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

func formatTstamp(t blip) string {
	μseconds := uint64(t) / 1000

	return fmt.Sprintf("TIMESTAMP 'epoch' + %d * INTERVAL '1 microseconds'",
		μseconds)
}

func storeInDb(l *list.List) bool {
	conn := pg.Connect(connString)
	defer pg.Close(conn)

	for item := range l.Iter() {
		execString := fmt.Sprintf("INSERT INTO blip (tstamp) VALUES %d",
			formatTstamp(item.(blip)))
		pg.Exec(conn, execString)
	}

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
