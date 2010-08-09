package main

import (
	"fmt"
	"time"
	"serial"
)

const (
	version = "v0.1" // Version of program
	inFlightMessages = 300 // How many blips do we allow to be in-flight?
)

type blip int64

func spawnFetcher() chan blip {
	c := make(chan blip, inFlightMessages)
	go func() {
		sp, err := serial.Open()
		if err != nil {
			panic(err)
		}

		for {
			sp.ReadLine()
			c <- blip(time.Nanoseconds())
		}
	}()

	return c
}

func main() {
	fmt.Printf("Blip storage daemon %s\n", version)
	fetchC := spawnFetcher()

	for {
		select {
		case x := <-fetchC:
			fmt.Printf("Read blip %s", x)
		default:
			fmt.Printf("Nothing to do")
		}
	}
}
