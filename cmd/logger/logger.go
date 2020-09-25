package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/fako1024/btscale/pkg/felicita"
	"github.com/fako1024/btscale/pkg/scale"
	"github.com/sirupsen/logrus"
)

type config struct {
	name string
	addr string
}

var log = logrus.New()

func main() {

	// Parse command line options
	var (
		cfg config
		s   scale.Scale
		err error
	)

	flag.StringVar(&cfg.name, "name", "FELICITA", "name of remote peripheral")
	flag.StringVar(&cfg.addr, "addr", "", "address of remote peripheral (MAC on Linux, UUID on OS X)")
	flag.Parse()

	s, err = felicita.New()
	if err != nil {
		log.Fatalf("Failed to initialize Felicita scale: %s", err)
	}

	s.SetDataHandler(func(data scale.DataPoint) {
		log.Warnf("Read DATA from Handler: %v, %v, %v, %v, %v", data, s.ConnectionStatus(), s.BatteryLevel(), s.IsBuzzingOnTouch(), s.ElapsedTime())
	})

	dataChan := make(chan scale.DataPoint, 256)
	s.SetDataChannel(dataChan)

	stateChan := make(chan scale.ConnectionStatus)
	s.SetStateChangeChannel(stateChan)

	go func() {
		for st := range stateChan {
			log.Warnf("State change: %v", st)
		}
	}()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		log.Infof("Got signal, terminating connection to device")
		s.Close()
		os.Exit(0)
	}()

	for {
		select {
		default:
			log.Warnf("Read DATA from Channel: %v, %v, %v, %v, %v", <-dataChan, s.ConnectionStatus(), s.BatteryLevel(), s.IsBuzzingOnTouch(), s.ElapsedTime())
		}
	}
}
