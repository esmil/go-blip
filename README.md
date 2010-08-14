# Go Blip - A powermeter monitor

In Labitat, we have a powermeter with a diode. This diode makes a blink whenever 1/1000 of a Kilowatthour has been consumed. So we hooked up an Arduino with a sensor on the powermeter to track its usage. The arduino code communicates on the serial port with a PC, so we can long-term store the blips.

Go-blip is a Go-program carrying out the storage-part of system, by writing each blip into a database. One goroutine is responsible for monitoring of the serial port and another is responsible for DB storage.

## TODO in no particular order

  * The whole code base needs some testing to make sure it is correct.
  * use the "syslog" package to log trouble
  * allow the program to daemonize
  * If we can't get a DB connection in 2 attempts, panic the program
    It is better to exit loudly than to try to survive.


