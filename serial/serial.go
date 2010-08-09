package serial

import (
	"os"
	"time"
)

type Serial bool

func Open() (Serial, os.Error) {
	return true, nil
}

func (s Serial) ReadLine() string {
	time.Sleep(3000000000)
	return "blip"
}
