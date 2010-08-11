package main

import (
	"container/list"
	"fmt"
	"os"
	"pgsql"
	"serial"
	"time"
)

const (
	version          = "v0.1" // Version of program
	inFlightMessages = 300    // How many blips do we allow to be in-flight?
	defaultSleeptime = 60 * 1000000000
	monitorDevice    = "/dev/tty.usbserial"
)

var (
	connParams       = &pgsql.ConnParams {
	                         Host: "127.0.0.1",
	                         Database: "blip",
	                         User: "foo",
	                         Password: "bar",
	                    }

)

type blip int64

func spawnFetcher() chan blip {
	c := make(chan blip, inFlightMessages)
	go func() {
		sp, err := serial.Open(monitorDevice, os.O_RDONLY, 0, serial.B9600_8E2)
		if err != nil {
			panic(err)
		}
		defer sp.Close()

		b := make([]byte, 1)
		for {
			sp.Read(b)
			c <- blip(time.Nanoseconds())
		}
	}()

	return c
}

func tstamp(t blip) uint64 {
	μseconds := uint64(t) / 1000
	return μseconds
}

func storeInDb(l *list.List) bool {
	// Move this to top-level
	conn, err := pgsql.Connect(connParams)
	if err != nil {
		return false
	}
	defer conn.Close()

	command := "INSERT INTO blip (tstamp) VALUES " +
		"TIMESTAMP 'epoch' + @ms * INTERVAL '1 microseconds'"

	tsParam := pgsql.NewParameter("@ms", pgsql.Integer)
	stmt, err := conn.Prepare(command, tsParam)
	if err != nil {
		// Something went wrong, try again later
		return false
	}
	defer stmt.Close()

	for l.Len() > 0 {
		elem := l.Front()
		item := elem.Value.(blip)
		tsParam.SetValue(tstamp(item))
		n, err := stmt.Execute()
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
