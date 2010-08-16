package main

import (
	"bufio"
	"container/list"
	"flag"
	"fmt"
	"log"
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
	logFile  = flag.String("logfile", "", "What file to log in, if any")
	logger *log.Logger
)

type blip int64

func spawnFetcher() chan blip {
	c := make(chan blip, inFlightMessages)
	logger.Log("Spawning the process responsible for serial fetching\n")
	go func() {
		sp, err := serial.Open(monitorDevice, os.O_RDONLY, 0, serial.B9600_8E2)
		if err != nil {
			panic(err)
		}
		defer sp.Close()
		logger.Log("Serial line successfully opened\n")

		buf := bufio.NewReader(sp)
		pr := textproto.NewReader(buf)
		for {
			ln, _ := pr.ReadLine()
			logger.Logf("Read line from serial port: %s\n", ln)
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
	connStr := fmt.Sprintf("host='%s' dbname='%s' user='%s' password='%s'",
		*host, *database, *user, *passwd)
	conn, err := pgsql.Connect(connStr, pgsql.LogError)
	conn.LogLevel = pgsql.LogVerbose
	if err != nil {
		return false
	}
	defer conn.Close()
	logger.Log("Successfully connected to DB")

	command := "INSERT INTO blip (tstamp) VALUES " +
		"(TIMESTAMP 'epoch' + %d * INTERVAL '1 microseconds')"

	for l.Len() > 0 {
		logger.Log("Storing blip")
		elem := l.Front()
		item := elem.Value.(blip)
		logger.Logf("Item is %d", item)
		n, err := conn.Execute(fmt.Sprintf(command, tstamp(item)))
		logger.Log("Stmt executed")
		if err != nil {
			logger.Logf("Problem executing statement: %s\n", err)
			return false
		}

		if n != 1 {
			panic("More than one row altered by insert")
		}

		l.Remove(elem)
	}
	logger.Log("All blips successfully stored\n")
	return true
}

func main() {
	flag.Parse()

	if *host == "" {
		fmt.Fprintf(os.Stderr, "Postgres host not defined\n")
		os.Exit(1)
	}

	logArg := os.Stdout
	if *logFile != "" {
		l, err := os.Open(*logFile, os.O_WRONLY | os.O_APPEND | os.O_CREAT,
			0660)
		if err != nil {
			panic("Logfile problem: " + err.String())
		}

		logArg = l
	}

	logger = log.New(logArg, nil, "blip ", log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	logger.Logf("Blip storage daemon %s\n", version)


	// Carry out a test connect to the Database early on
	//   Saves us the hassle if it goes wrong later on
	connStr := fmt.Sprintf("host='%s' dbname='%s' user='%s' password='%s'",
		*host, *database, *user, *passwd)
	logger.Logf("Making testconnection to pgsql://%s/%s\n", *host, *database)
	db, err := pgsql.Connect(connStr, pgsql.LogError)
	if err != nil {
		panic(err)
	}
	logger.Log("Connection successful")
	db.Close()


	fetchC := spawnFetcher()
	l := list.New()

	logger.Log("Entering storage processor loop\n")
	for {
		select {
		case x := <-fetchC:
			l.PushBack(x)
		default:
			if l.Len() > 0 {
				logger.Logf("Read %d blips, storing\n", l.Len())
				r := storeInDb(l)
				if r == false {
					logger.Log("Warning: DB has problems\n")
				}
			}
			time.Sleep(defaultSleeptime)
		}
	}
}
