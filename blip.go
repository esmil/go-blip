package main

import (
	"container/list"
	"fmt"
	"os"
	"pgsql"
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
		defer sp.Close()
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
	// Move this to top-level
	connParams := &pgsql.ConnParams {
	Host: "127.0.0.1",
	Database: "blip",
	User: "foo",
	Password: "bar",
	}
	conn, err := pgsql.Connect(connParams)
	if err != nil {
		return false
	}
	defer conn.Close()

	for l.Len() > 0 {
		elem := l.Front()
		item := elem.Value
		execString := fmt.Sprintf("INSERT INTO blip (tstamp) VALUES %d",
			formatTstamp(item.(blip)))
		n, err := conn.Execute(execString)
		if err != nil {
			// Something went wrong, try again later
			return false
		}

		if n != 1 {
			panic("More than one row altered by insert")
		}

		l.Remove(elem)
	}

	return true
}

func main() {
	fmt.Printf("Blip storage daemon %s\n", version)
	fetchC := spawnFetcher()
	l := list.New()

	for {
		for x := range fetchC {
			l.PushBack(x)
			fmt.Printf("Read blip %s", x)
		}

		r := storeInDb(l)
		if r == false {
			fmt.Printf("Warning: DB has problems\n")
		}
		time.Sleep(defaultSleeptime)
	}
}
