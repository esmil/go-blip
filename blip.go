package main

import (
	"bufio"
	"container/list"
	"flag"
	"fmt"
	"net/textproto"
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
	host     = flag.String("host", "", "Postgres database host")
	database = flag.String("db", "blip", "Name of the postgres database to connect to")
	user     = flag.String("user", "blip", "Login for the database")
	passwd   = flag.String("passwd", "", "Password for the database")
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

		buf := bufio.NewReader(sp)
		pr := textproto.NewReader(buf)
		for {
			pr.ReadLine()
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
	connParams := &pgsql.ConnParams{
		Host:     *host,
		Database: *database,
		User:     *user,
		Password: *passwd,
	}
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
	flag.Parse()

	if *host == "" {
		fmt.Fprintf(os.Stderr, "Postgres host not defined")
		os.Exit(1)
	}

	// Carry out a test connect to the Database early on
	//   Saves us the hassle if it goes wrong later on
	connParams := &pgsql.ConnParams{
		Host:     *host,
		Database: *database,
		User:     *user,
		Password: *passwd,
	}

	db, err := pgsql.Connect(connParams)
	if err != nil {
		panic(err)
	}
	db.Close()

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
