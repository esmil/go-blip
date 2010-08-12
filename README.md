# Go Blip - A powermeter monitor

In Labitat, we have a powermeter with a diode. This diode makes a blink whenever 1/1000 of a Kilowatthour has been consumed. So we hooked up an Arduino with a sensor on the powermeter to track its usage. The arduino code communicates on the serial port with a PC, so we can long-term store the blips.

Go-blip is a Go-program carrying out the storage-part of system, by writing each blip into a database. One goroutine is responsible for monitoring of the serial port and another is responsible for DB storage.

## TODO in no particular order

  * The whole code base needs some testing to make sure it is correct.
  * use the "log" package to log to a file
  * use the "syslog" package to log trouble
  * allow the program to daemonize
  * Do more testing and integration of the go-pgsql package.

