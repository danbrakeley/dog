package main

import (
	"flag"
	"math/rand"
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
	log.SetMinLevel(dog.Transient)

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

	for {
		switch rand.Intn(10) {
		case 0:
			msg := "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum."
			switch rand.Intn(10) {
			case 0:
				log.Warning(msg)
			case 1:
				log.Error(msg)
			case 2:
				log.Verbose(msg)
			default:
				log.Info(msg)
			}
		case 1:
			log.Info("What Is The Answer and Why?")
		case 2:
			log.Info("I Blame Your Mother")
		case 3:
			log.Info("What Are The Civilian Applications?")
		case 4:
			log.Info("Clear Air Turbulence")
		case 5:
			log.Info("Fate Amenable To Change")
		case 6:
			log.Info("Frank Exchange Of Views")
		case 7:
			log.Info("Anticipation Of A New Lover's Arrival, The")
		case 8:
			log.Warning("Something exciting may have happened? Then again, maybe not.")
		case 9:
			log.Error("a lovely day at the beach")
		}
		time.Sleep(time.Duration(rand.Intn(2000)+200) * time.Millisecond)
	}

	// log.Fatal("this is a Fatal log line")
}
