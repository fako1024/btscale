package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/fako1024/btscale/pkg/felicita"
	"github.com/fako1024/btscale/pkg/scale"
)

type config struct {
	name string
	addr string
}

func main() {

	logger := scale.NewDefaultLogger(false)

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
		logger.Fatalf("failed to initialize Felicita scale: %s", err)
	}

	s.SetDataHandler(func(data scale.DataPoint) {
		logger.Warnf("read DATA from Handler: %v, %v, %v, %v, %v", data, s.ConnectionStatus(), s.BatteryLevel(), s.IsBuzzingOnTouch(), s.ElapsedTime())
	})

	dataChan := make(chan scale.DataPoint, 256)
	s.SetDataChannel(dataChan)

	stateChan := make(chan scale.ConnectionStatus)
	s.SetStateChangeChannel(stateChan)

	go func() {
		for st := range stateChan {
			logger.Warnf("state change: %v", st)
		}
	}()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGTERM)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		logger.Infof("got signal, terminating connection to device")
		if err := s.Close(); err != nil {
			logger.Errorf("failed to close device: %s", err)
		}
		os.Exit(0)
	}()

	for v := range dataChan {
		logger.Warnf("read DATA from Channel: %v, %v, %v, %v, %v", v, s.ConnectionStatus(), s.BatteryLevel(), s.IsBuzzingOnTouch(), s.ElapsedTime())
	}
}
