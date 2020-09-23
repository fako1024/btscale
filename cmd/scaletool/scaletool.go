package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/fako1024/btscale/pkg/felicita"
	"github.com/fako1024/btscale/pkg/scale"
	"github.com/sirupsen/logrus"
)

type config struct {
	name string
	addr string

	togglePrecision bool
	toggleBuzzer    bool
}

var log = logrus.New()

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	// Parse command line options
	var (
		cfg config
		s   scale.Scale
		err error
	)

	flag.StringVar(&cfg.name, "name", "FELICITA", "Name of remote peripheral")
	flag.StringVar(&cfg.addr, "addr", "", "Address of remote peripheral (MAC on Linux, UUID on OS X)")

	flag.BoolVar(&cfg.togglePrecision, "p", false, "Toggle the scale precision")
	flag.BoolVar(&cfg.toggleBuzzer, "b", false, "Toggle the buzzer on touch / action feature")
	flag.Parse()

	s, err = felicita.New()
	if err != nil {
		return fmt.Errorf("Failed to initialize Felicita scale: %s", err)
	}
	defer s.Close()

	for {
		time.Sleep(time.Second)
		if s.ConnectionStatus().State == scale.StateConnected {
			break
		}
	}

	if cfg.togglePrecision {
		if err := s.TogglePrecision(); err != nil {
			return fmt.Errorf("Failed to toggle scale precision: %s", err)
		}
	}
	if cfg.toggleBuzzer {
		if err := s.ToggleBuzzingOnTouch(); err != nil {
			return fmt.Errorf("Failed to toggle buzzer on touch / action: %s", err)
		}
	}

	return nil
}
