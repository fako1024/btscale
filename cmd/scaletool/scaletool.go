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

	togglePrecision bool
	toggleBuzzer    bool
}

var log = logrus.New()

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (err error) {

	// Parse command line options
	var (
		cfg config
		s   scale.Scale
	)

	flag.StringVar(&cfg.name, "name", "FELICITA", "Name of remote peripheral")

	flag.BoolVar(&cfg.togglePrecision, "p", false, "Toggle the scale precision")
	flag.BoolVar(&cfg.toggleBuzzer, "b", false, "Toggle the buzzer on touch / action feature")
	flag.Parse()

	s, err = felicita.New()
	if err != nil {
		return fmt.Errorf("failed to initialize Felicita scale: %s", err)
	}
	defer func() {
		if cerr := s.Close(); cerr != nil {
			err = cerr
			return
		}
	}()

	for {
		time.Sleep(time.Second)
		if s.ConnectionStatus().State == scale.StateConnected {
			break
		}
	}

	if cfg.togglePrecision {
		if err := s.TogglePrecision(); err != nil {
			return fmt.Errorf("failed to toggle scale precision: %s", err)
		}
	}
	if cfg.toggleBuzzer {
		if err := s.ToggleBuzzingOnTouch(); err != nil {
			return fmt.Errorf("failed to toggle buzzer on touch / action: %s", err)
		}
	}

	return nil
}
