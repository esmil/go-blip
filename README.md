# Go Blip - A powermeter monitor

In Labitat, we have a powermeter with a diode. This diode makes a blink whenever 1/1000 of a Kilowatthour has been consumed. So we hooked up an Arduino with a sensor on the powermeter to track its usage. The arduino code communicates on the serial port with a PC, so we can long-term store the blips.

Go-blip is a Go-program carrying out the storage-part of system, by writing each blip into a database. One goroutine is responsible for monitoring of the serial port and another is responsible for DB storage.

## TODO in no particular order

  * The whole code base needs some testing to make sure it is correct.
  * Set termios parameters correctly
  * use the "flag" module to parse and set command-line options
  * use the "syslog" package to log trouble
  * allow the program to daemonize
  * make sure all needed packages can goinstall
  * The go-termios project need some TLC:
     * Remember to reset to the old termios settings when closing the port.
     * Improve the ability to set parameters to termios
  * The go-pg project need some TLC:
     * All the unsafe.Pointer passing business should be packed up
     * Type encapsulation is key
     * Make calls into interface-calls
     * Look at what other SQL-db interfaces are doing and mimic them.
     * Take a close look at [Ross Cox sqlite3 binding code](http://code.google.com/p/gosqlite/source/browse/sqlite/sqlite.go)
     * handle errors correctly by returning tuples of (x, err)
