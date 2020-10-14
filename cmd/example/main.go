package main

import (
	"flag"
	"time"

	"github.com/danbrakeley/dog"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()

	log, err := dog.Create(*addr)
	if err != nil {
		panic(err)
	}
	defer log.Close()

	log.Info("first")
	time.Sleep(time.Second)
	log.Info("second")
	time.Sleep(time.Second)
	log.Info("third")
	time.Sleep(time.Second)
	log.Transient("this is a Transient log line")
	log.Verbose("this is a Verbose log line")
	log.Info("this is an Info log line")
	log.Warning("this is a Warning log line")
	log.Error("this is an Error log line")
	time.Sleep(time.Second)

	log.Info("This is my application doing some stuff that is important!")
	log.Info("This is my application doing some stuff that is important!")
	log.Warning("Something exciting may have happened? Then again, maybe not.")
	time.Sleep(1 * time.Second)
	log.Info("one line followed quickly by")
	log.Info("a second line")
	log.Error("well fuck")
	log.Fatal("this is a Fatal log line")
}
